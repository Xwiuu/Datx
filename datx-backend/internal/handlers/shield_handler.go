package handlers

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/xwiuu/datx-backend/internal/database"
	"github.com/xwiuu/datx-backend/internal/models"
	"gorm.io/gorm"
)

// Ofuscador Elite: Transforma o JS legível em um payload Base64 injetável e auto-executável
func obfuscateJS(rawJS string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(rawJS))
	return fmt.Sprintf(`(function(){
		try {
			var _0x1a2b = atob("%s");
			var _0x3c4d = document.createElement('script');
			_0x3c4d.type = 'text/javascript';
			_0x3c4d.text = _0x1a2b;
			document.head.appendChild(_0x3c4d);
			document.head.removeChild(_0x3c4d);
		} catch(e) {}
	})();`, encoded)
}

// RegisterTrace: A rota secreta que o script JS chama para atualizar os Big Numbers (Analytics)
func RegisterTrace(c *fiber.Ctx) error {
	slug := c.Params("slug")
	type TraceReq struct {
		Type   string `json:"type"`   // "human" ou "bot"
		Device string `json:"device"` // "mobile" ou "pc"
	}
	var req TraceReq
	if err := c.BodyParser(&req); err != nil {
		return c.SendStatus(400)
	}

	// Incrementa os contadores no banco de dados de forma atômica (Thread-safe)
	update := database.DB.Model(&models.Link{}).Where("slug = ?", slug)
	update.UpdateColumn("total_visits", gorm.Expr("total_visits + ?", 1))

	if req.Type == "human" {
		update.UpdateColumn("human_visits", gorm.Expr("human_visits + ?", 1))
	} else {
		update.UpdateColumn("bot_blocked", gorm.Expr("bot_blocked + ?", 1))
	}

	if req.Device == "mobile" {
		update.UpdateColumn("mobile_visits", gorm.Expr("mobile_visits + ?", 1))
	} else {
		update.UpdateColumn("pc_visits", gorm.Expr("pc_visits + ?", 1))
	}

	return c.SendStatus(200)
}

// ServeShieldScript: O motor principal que entrega o script de defesa customizado
func ServeShieldScript(c *fiber.Ctx) error {
	slug := c.Params("slug")
	slug = strings.TrimSuffix(slug, ".js")

	var link models.Link
	if err := database.DB.Where("slug = ? AND status = 'active'", slug).First(&link).Error; err != nil {
		c.Set("Content-Type", "application/javascript")
		return c.SendString("console.log('DATX Shield: Node inativo.');")
	}

	var js strings.Builder

	// --- INÍCIO DA CONSTRUÇÃO DO JAVASCRIPT ---
	js.WriteString("(function() {\n")
	js.WriteString(fmt.Sprintf("  const _r = '%s';\n", link.RedirectURL))
	js.WriteString(fmt.Sprintf("  const _s = '%s';\n", slug))

	// Função Interna de Rastreio (Beacon) para alimentar o Analytics
	js.WriteString(`
	const _t = (type) => {
		fetch('http://localhost:8080/v1/shield/trace/' + _s, {
			method: 'POST',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({
				type: type,
				device: /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) ? 'mobile' : 'pc'
			})
		});
	};
	`)

	// 🛑 Iframe Buster: Mata ferramentas de AdSpy (AdHeart, SpyOver, etc)
	if link.IframeBuster {
		js.WriteString(`
		if (window.top !== window.self) { 
			_t('bot'); window.top.location.replace(_r); return; 
		}
		`)
	}

	// 📱 Controle de Dispositivo (Mobile vs PC)
	js.WriteString(`const _isM = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);\n`)
	if !link.AllowMobile {
		js.WriteString("if (_isM) { _t('bot'); window.location.replace(_r); return; }\n")
	}
	if !link.AllowPC {
		js.WriteString("if (!_isM) { _t('bot'); window.location.replace(_r); return; }\n")
	}

	// 🤖 Bloqueio de Bots & Fingerprint (Hardware + Comportamento)
	if link.BlockBots {
		js.WriteString(`
		const _bl = ['bot', 'crawler', 'spider', 'headless', 'phantom', 'selenium', 'puppeteer', 'facebookexternalhit', 'lighthouse', 'chrome-lighthouse'];
		const _ua = navigator.userAgent.toLowerCase();
		if (_bl.some(k => _ua.includes(k)) || navigator.webdriver) {
			_t('bot'); window.location.replace(_r); return;
		}
		
		// Prova de Humanidade (O Matador de Robôs)
		let _hum = false;
		const _validate = () => { if(!_hum) { _hum = true; _t('human'); } };
		['mousemove', 'touchstart', 'scroll', 'keydown', 'click'].forEach(e => document.addEventListener(e, _validate, {once:true}));
		
		// Prazo: 4 Segundos para provar atividade humana
		setTimeout(() => {
			if (!_hum) {
				_t('bot');
				document.body.innerHTML = '<div style="background:#000;width:100vw;height:100vh;"></div>';
				window.location.replace(_r);
			}
		}, 4000);
		`)
	}

	// 🔒 Anti-Inspect (F12, Botão Direito e Debugger Loop)
	if link.AntiInspect {
		js.WriteString(`
		document.addEventListener('contextmenu', e => e.preventDefault());
		document.addEventListener('keydown', e => {
			if (e.keyCode == 123 || (e.ctrlKey && e.shiftKey && (e.keyCode == 73 || e.keyCode == 74 || e.keyCode == 67)) || (e.ctrlKey && e.keyCode == 85)) {
				e.preventDefault(); return false;
			}
		});
		setInterval(() => {
			const st = performance.now(); debugger; const et = performance.now();
			if (et - st > 100) { _t('bot'); window.location.replace(_r); }
		}, 500);
		`)
	}

	// 🪃 Boomerang (Anti-Clone)
	if link.AntiClone {
		js.WriteString(fmt.Sprintf(`
		if (window.location.hostname !== "localhost" && window.location.hostname !== "127.0.0.1") {
			try {
				const _target = new URL("%s").hostname;
				if (window.location.hostname !== _target) {
					window.location.replace("%s"); return;
				}
			} catch(e) {}
		}
		`, link.PageURL, link.PageURL))
	}

	// Se for Humano (Adicione isso na parte do código que libera o acesso)
	js.WriteString(fmt.Sprintf("fetch('http://localhost:8080/v1/shield/event/%s?type=human');\n", slug))

	// Se for Bot (Adicione isso na parte que redireciona ou bloqueia)
	js.WriteString(fmt.Sprintf("fetch('http://localhost:8080/v1/shield/event/%s?type=bot');\n", slug))
	// --- FIM DO CÓDIGO PURO ---

	js.WriteString("})();\n")

	// Ofusca o código final para segurança máxima
	finalPayload := obfuscateJS(js.String())

	c.Set("Content-Type", "application/javascript")
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate")
	return c.SendString(finalPayload)

}
