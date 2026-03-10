package cloaker

import (
	"strings"
)

type HardwareInfo struct {
	GPU   string `json:"gpu"`
	Mem   int    `json:"mem"`
	Cores int    `json:"cores"`
	Res   string `json:"res"`
}

func IsFakeHardware(info HardwareInfo) bool {
	// 1. Blacklist de Renderizadores de Software (Típicos de BOTS/DataCenters)
	fakeRenderers := []string{
		"SwiftShader",
		"llvmpipe",
		"VirtualBox",
		"VMware",
		"Microsoft Basic Render Driver",
		"Headless",
	}

	gpuLower := strings.ToLower(info.GPU)
	for _, fake := range fakeRenderers {
		if strings.Contains(gpuLower, strings.ToLower(fake)) {
			return true // É um BOT fingindo ser PC
		}
	}

	// 2. Consistência: Se não tem memória ou núcleos, é suspeito (exceto navegadores muito antigos)
	if info.Mem == 0 && info.Cores == 0 {
		return true
	}

	return false
}
