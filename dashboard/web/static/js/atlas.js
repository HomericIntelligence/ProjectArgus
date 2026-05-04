// Atlas UI — htmx + SSE scaffolding
document.addEventListener("DOMContentLoaded", function () {
    const dot = document.querySelector(".conn-dot");

    // SSE connection placeholder — wired when /events endpoint lands (M2)
    if (typeof EventSource !== "undefined" && dot) {
        const src = new EventSource("/events");
        src.onopen = () => dot.classList.add("connected");
        src.onerror = () => dot.classList.remove("connected");
    }
});
