package cloaker

import "fmt"

// GetAntiCloneScript gera o script de proteção injetável
func GetAntiCloneScript(allowedDomain string, boomerangCheckoutURL string) string {
	return fmt.Sprintf(`
	<script>
		if (window.location.hostname !== "%s" && window.location.hostname !== "localhost") {
			// A Vingança do Boomerang: O cara clonou? O tráfego vai pro seu bolso!
			window.location.href = "%s"; 
		}

		// 2. Bloqueio de Inspeção
		document.addEventListener('contextmenu', e => e.preventDefault());
		document.onkeydown = function(e) {
			if(e.keyCode == 123 || (e.ctrlKey && e.shiftKey && (e.keyCode == 'I'.charCodeAt(0) || e.keyCode == 'J'.charCodeAt(0))) || (e.ctrlKey && e.keyCode == 'U'.charCodeAt(0))) {
				return false;
			}
		};

		// 3. Debugger Loop (Trava se abrir o console)
		setInterval(function() {
			(function(c) {
				(function(a) {
					if (a === "custom") {
						(function b(i) {
							if (("" + i / i).length !== 1 || i %% 20 === 0) {
								(function() {}).constructor("debugger")();
							} else {
								(function() {}).constructor("debugger")();
							}
							b(++i);
						})(0);
					}
				})(c);
			})("custom");
		}, 1000);
	})();
	</script>
	`, allowedDomain, boomerangCheckoutURL)
}
