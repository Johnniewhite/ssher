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

const year = document.getElementById("year");
if (year) year.textContent = new Date().getFullYear();

const menuButton = document.querySelector(".menu-button");
const navLinks = document.getElementById("nav-links");
menuButton?.addEventListener("click", () => {
  const open = navLinks?.classList.toggle("open") ?? false;
  menuButton.setAttribute("aria-expanded", String(open));
  menuButton.setAttribute("aria-label", open ? "Close navigation" : "Open navigation");
});
navLinks?.querySelectorAll("a").forEach((link) => link.addEventListener("click", () => {
  navLinks.classList.remove("open");
  menuButton?.setAttribute("aria-expanded", "false");
}));

document.querySelectorAll(".code-block button").forEach((button) => {
  button.addEventListener("click", async () => {
    const code = button.parentElement?.querySelector("pre")?.textContent ?? "";
    try {
      await navigator.clipboard.writeText(code.trim());
      button.textContent = "Copied";
      window.setTimeout(() => { button.textContent = "Copy"; }, 1400);
    } catch {
      button.textContent = "Select code";
    }
  });
});

const docsSearch = document.getElementById("docs-search");
docsSearch?.addEventListener("input", () => {
  const query = docsSearch.value.toLowerCase().trim();
  document.querySelectorAll(".guide-card").forEach((card) => {
    card.hidden = query !== "" && !card.textContent.toLowerCase().includes(query);
  });
  document.querySelectorAll(".guide-group").forEach((group) => {
    group.hidden = !group.querySelector(".guide-card:not([hidden])");
  });
});
