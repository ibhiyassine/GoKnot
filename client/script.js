let proxyUrl = "";
let autoInterval = null;
let requestCount = 0;

// 1. Initialize: Read Config on Load
window.onload = async () => {
    const statusEl = document.getElementById("config-status");
    const proxyEl = document.getElementById("proxy-url");

    try {
        // We look for config.json in the parent folder relative to this file
        // IMPORTANT: We must run the server from the Project ROOT for this to work
        const response = await fetch('/config.json');
        
        if (!response.ok) throw new Error("Config not found");
        
        const config = await response.json();
        
        // Construct the URL based on the port in config
        // Assuming localhost, but you could add a "host" field to config later
        proxyUrl = `http://localhost:${config.port}`;
        
        statusEl.innerText = "✅ Config Loaded";
        statusEl.style.color = "#2ecc71";
        proxyEl.innerText = `Target: ${proxyUrl}`;
        
    } catch (err) {
        statusEl.innerText = "❌ Failed to load /config.json";
        statusEl.style.color = "#e74c3c";
        console.error(err);
        appendLog("System", "Error: Could not load config.json. Are you running the server from Project Root?", "error");
    }
};

// 2. Core Request Function
async function sendRequest() {
    const startTime = performance.now();
    try {
        const response = await fetch(proxyUrl);
        const text = await response.text();
        const endTime = performance.now();

        // Update Stats
        requestCount++;
        document.getElementById("req-count").innerText = requestCount;
        document.getElementById("latency").innerText = Math.round(endTime - startTime) + "ms";

        appendLog("Proxy", text, "default");

    } catch (err) {
        console.error(err)
        appendLog("Error", "Connection Failed (Is Proxy Running?)", "error");
    }
}

function clearLog() {
    document.getElementById("log-window").innerHTML = "";
    requestCount = 0;
    document.getElementById("req-count").innerText = "0";
}

function appendLog(source, message, type) {
    const logWindow = document.getElementById("log-window");
    const entry = document.createElement("div");
    const time = new Date().toLocaleTimeString();
    
    entry.className = "log-entry";
    entry.innerHTML = `<span style="color:#777">[${time}]</span> <span class="${type}">${message}</span>`;
    
    // Add to top
    logWindow.prepend(entry);
}
