const installTabs = document.querySelectorAll(".install-tabs button");
const installCommand = document.getElementById("install-command");
const copyButton = document.querySelector(".copy-button");

for (const tab of installTabs) {
  tab.addEventListener("click", () => {
    for (const item of installTabs) {
      const selected = item === tab;
      item.classList.toggle("active", selected);
      item.setAttribute("aria-selected", String(selected));
    }
    installCommand.textContent = tab.dataset.command;
    copyButton.textContent = "Copy";
  });
}

copyButton?.addEventListener("click", async () => {
  try {
    await navigator.clipboard.writeText(installCommand.textContent);
    copyButton.textContent = "Copied";
  } catch {
    copyButton.textContent = "Select";
    const selection = window.getSelection();
    const range = document.createRange();
    range.selectNodeContents(installCommand);
    selection.removeAllRanges();
    selection.addRange(range);
  }
});

document.getElementById("year").textContent = new Date().getFullYear();
