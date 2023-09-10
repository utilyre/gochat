const form = document.querySelector("form");

document.body.addEventListener("htmx:wsAfterSend", () => form.reset());
