// import OSC from "./osc.min.js";

function main(){
    "use strict";

    const plugin = new OSC.WebsocketClientPlugin({ host: location.hostname, port: location.port });
    const osc = new OSC({ plugin: plugin });

    osc.on('open',     ()          => { statusLabel.innerHTML = "Server Running"; });
    osc.on('/name',    (message)   => procName(message));
    osc.on('/message', (message)   => procMsg(message));
    osc.on('/time',    (message)   => procTime(message));
    osc.on('/clock',   (message)   => procClock(message));
    osc.on('/tminus',  (message)   => procTminus(message));
    osc.on('/reset',   ()          => resetDisplay());
    osc.on('/refresh', ()          => location.reload());
    osc.on('/connect', ()          => resetDisplay());
    osc.on('/stop',    ()          => plugin.close());
    osc.on('close',    ()          => closeDisplay());

    const timecodeHours    = document.getElementById("timecode-hours");
    const timecodeMinutes  = document.getElementById("timecode-minutes");
    const timecodeSeconds  = document.getElementById("timecode-seconds");
    const timecodeMS       = document.getElementById("timecode-ms");
    const timecodeClipName = document.getElementById("clipname");
    const table            = document.getElementById("table");
    const tableBorder      = document.getElementById("tableborder");
    const clipLengthLabel  = document.getElementById("ms");
    const statusLabel      = document.getElementById("status");
    const messageEl        = document.getElementById("msg");

    // Interpolation state
    let remainingMs  = null;
    let receivedAt   = null;
    let isCounting   = true;   // true = counting down (T-), false = counting up (T+)
    let rafId        = null;


    resetDisplay();
    osc.open();

    // ── display helpers ──────────────────────────────────────────────

    function setVisible(elems, visible) {
        for (let i = 0; i < elems.length; i++) {
            elems[i].classList.toggle("clock-hidden", !visible);
        }
    }

    function renderDigits(h, min, sec, ms) {
        timecodeHours.innerHTML   = String(h).padStart(2, "0");
        timecodeMinutes.innerHTML = String(min).padStart(2, "0");
        timecodeSeconds.innerHTML = String(sec).padStart(2, "0");
        timecodeMS.innerHTML      = String(ms).padStart(3, "0");
    }

    function setAlert(on) {
        const color = on ? "#ff4545" : "#45ff45";
        table.style.color = color;
        if (tableBorder) tableBorder.style.borderColor = on ? "#ff4545" : "#4b5457";
    }

    // ── OSC handlers ─────────────────────────────────────────────────

    function closeDisplay() {
        statusLabel.innerHTML = "Server Stopped";
        stopTick();
        remainingMs = null;
        receivedAt  = null;
        renderDigits(0, 0, 0, 0);
        setAlert(false);
        clipLengthLabel.innerHTML = "0.000s";
    }

    function resetDisplay() {
        stopTick();
        remainingMs = null;
        receivedAt  = null;
        renderDigits(0, 0, 0, 0);
        setAlert(false);
        clipLengthLabel.innerHTML = "0.000s";
    }

    function procName(data) {
        timecodeClipName.innerHTML = data.args[0];
    }

    function procTminus(data) {
        isCounting = data.args[0] === true;
        const minusEls = document.getElementsByClassName("minus");
        if (minusEls[0]) minusEls[0].innerHTML = isCounting ? "-" : "+";
    }

    function procClock(data) {
        const showHours     = data.args[0];
        const showMsVal     = data.args[1];
        const showSign      = data.args[2];
        const clipNameSizeV = data.args[3];

        setVisible(document.getElementsByClassName("hours"), showHours);
        setVisible(document.getElementsByClassName("ms"),    showMsVal);
        setVisible(document.getElementsByClassName("minus"), showSign);

        if (clipNameSizeV != null) {
            const wrap = document.getElementById("clipname-wrap");
            if (wrap) wrap.style.fontSize = clipNameSizeV + "vw";
        }
    }

    async function procMsg(data) {
        const text = data.args[0];
        messageEl.innerHTML    = (text === "") ? "" : text;
        messageEl.style.display = (text === "") ? "none" : "block";
        if (text === "") return;
        for (let i = 0; i < 3; i++) {
            messageEl.style.color = "#ff4545";
            await new Promise(r => setTimeout(r, 500));
            messageEl.style.color = "#FDFBF7";
            await new Promise(r => setTimeout(r, 500));
        }
    }

    function procTime(data) {
        clipLengthLabel.innerHTML = data.args[1];

        // Parse "-HH:MM:SS.mmm"
        const str   = data.args[0];
        const parts = str.split(":");
        const h  = parseInt(parts[0].replace(/[^0-9]/g, ""), 10);
        const m  = parseInt(parts[1], 10);
        const ss = parts[2].split(".");
        const s  = parseInt(ss[0], 10);
        const ms = parseInt(ss[1], 10);

        remainingMs = ((h * 3600 + m * 60 + s) * 1000) + ms;
        receivedAt  = performance.now();

        if (!rafId) tick();
    }

    // ── interpolation loop ────────────────────────────────────────────

    function stopTick() {
        if (rafId) {
            cancelAnimationFrame(rafId);
            rafId = null;
        }
    }

    function tick() {
        rafId = null; // will be reassigned if we continue

        if (remainingMs === null || receivedAt === null) return;

        const elapsed = performance.now() - receivedAt;

        const current = isCounting
            ? Math.max(remainingMs - elapsed, 0)   // T- : count down
            : Math.min(remainingMs + elapsed, 359999999); // T+ : count up

        const totalMs = Math.round(current);
        const ms  = totalMs % 1000;
        const sec = Math.floor(totalMs / 1000) % 60;
        const min = Math.floor(totalMs / 60000) % 60;
        const h   = Math.floor(totalMs / 3600000);

        renderDigits(h, min, sec, ms);
        setAlert(h === 0 && min === 0 && sec <= 10);

        rafId = requestAnimationFrame(tick);
    }
}

main();
