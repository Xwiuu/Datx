(function () {
    const getHardwareInfo = () => {
        const canvas = document.createElement('canvas');
        const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
        if (!gl) return { gpu: 'none' };

        const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
        return {
            gpu: debugInfo ? gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) : 'unknown',
            mem: navigator.deviceMemory || 0,
            cores: navigator.hardwareConcurrency || 0,
            res: `${window.screen.width}x${window.screen.height}`
        };
    };

    const info = getHardwareInfo();
    // Envia para o servidor via Beacon (não atrasa o carregamento)
    const data = new FormData();
    data.append('info', JSON.stringify(info));
    navigator.sendBeacon('/v1/trace-hardware', data);
})();