(() => {
  const measurementId = "G-T2Q2L6RSTJ";
  const consentCookie = "ssher_analytics_consent";
  const consentMaxAge = 60 * 60 * 24 * 180;

  function readConsent() {
    const value = document.cookie
      .split("; ")
      .find((entry) => entry.startsWith(`${consentCookie}=`))
      ?.split("=")[1];
    return value === "granted" || value === "denied" ? value : null;
  }

  function writeConsent(value) {
    const domain = location.hostname === "getssher.com" || location.hostname.endsWith(".getssher.com")
      ? "; Domain=.getssher.com"
      : "";
    document.cookie = `${consentCookie}=${value}; Path=/; Max-Age=${consentMaxAge}; SameSite=Lax; Secure${domain}`;
  }

  window.dataLayer = window.dataLayer || [];
  function gtag() { window.dataLayer.push(arguments); }
  window.gtag = window.gtag || gtag;

  const storedConsent = readConsent();
  gtag("consent", "default", {
    analytics_storage: storedConsent === "granted" ? "granted" : "denied",
    ad_storage: "denied",
    ad_user_data: "denied",
    ad_personalization: "denied",
    wait_for_update: 500,
  });
  gtag("js", new Date());
  gtag("config", measurementId, {
    send_page_view: false,
    allow_google_signals: false,
    allow_ad_personalization_signals: false,
    cookie_flags: "SameSite=Lax;Secure",
  });

  const script = document.createElement("script");
  script.async = true;
  script.src = `https://www.googletagmanager.com/gtag/js?id=${measurementId}`;
  document.head.appendChild(script);

  function track(eventName, parameters = {}) {
    gtag("event", eventName, parameters);
  }

  function trackPageView() {
    track("page_view", {
      page_title: document.title,
      page_location: `${location.origin}${location.pathname}`,
      page_path: location.pathname,
      site_area: location.pathname.startsWith("/docs") ? "docs" : "marketing",
    });
  }

  function setConsent(value) {
    writeConsent(value);
    gtag("consent", "update", { analytics_storage: value });
    document.querySelector("[data-analytics-consent]")?.remove();
    track("analytics_consent_updated", { consent_state: value });
  }

  window.ssherAnalytics = { track, setConsent };

  function showConsent() {
    if (storedConsent) return;
    const banner = document.createElement("aside");
    banner.className = "analytics-consent";
    banner.dataset.analyticsConsent = "";
    banner.setAttribute("aria-label", "Analytics preferences");
    banner.innerHTML = `
      <div><strong>Useful analytics. No advertising.</strong><p>Help us understand which ssher pages and features are useful. We never send server details, commands, credentials, emails, or access-link tokens.</p></div>
      <div class="analytics-consent-actions"><a href="/docs/security/#website-analytics">Learn more</a><button type="button" data-consent="denied">No thanks</button><button type="button" data-consent="granted">Allow analytics</button></div>`;
    const style = document.createElement("style");
    style.textContent = `.analytics-consent{position:fixed;z-index:9999;right:18px;bottom:18px;width:min(520px,calc(100vw - 36px));display:flex;gap:20px;align-items:center;padding:18px 20px;background:#111;color:#fff;border:1px solid #353535;box-shadow:0 18px 55px rgba(0,0,0,.28);font-family:Inter,system-ui,sans-serif}.analytics-consent strong{font-size:14px}.analytics-consent p{margin:5px 0 0;color:#b8b8b8;font-size:12px;line-height:1.5}.analytics-consent-actions{display:flex;flex-shrink:0;gap:7px;align-items:center}.analytics-consent a,.analytics-consent button{font:600 11px Inter,system-ui,sans-serif}.analytics-consent a{color:#d1ff3f;margin-right:3px}.analytics-consent button{border:1px solid #444;background:#202020;color:#fff;padding:9px 11px;cursor:pointer}.analytics-consent button:last-child{background:#d1ff3f;border-color:#d1ff3f;color:#0b0b0b}@media(max-width:650px){.analytics-consent{left:12px;right:12px;bottom:12px;width:auto;display:block}.analytics-consent-actions{margin-top:14px;justify-content:flex-end;flex-wrap:wrap}}`;
    document.head.appendChild(style);
    document.body.appendChild(banner);
    banner.querySelectorAll("[data-consent]").forEach((button) => button.addEventListener("click", () => setConsent(button.dataset.consent)));
  }

  function setupAnalytics() {
    trackPageView();
    showConsent();

    document.addEventListener("click", (event) => {
      const target = event.target.closest("a, button");
      if (!target) return;
      if (target.matches(".install-tabs button")) {
        track("install_method_selected", { method: target.textContent.trim().toLowerCase() });
      } else if (target.matches(".copy-button")) {
        track("install_command_copied", { method: document.querySelector(".install-tabs .active")?.textContent.trim().toLowerCase() || "unknown" });
      } else if (target.matches('a[href="#install"], a[href="/#install"]')) {
        track("install_cta_clicked", { link_text: target.textContent.trim().slice(0, 80) });
      } else if (target instanceof HTMLAnchorElement && target.hostname === "cloud.getssher.com") {
        track("cloud_cta_clicked", { destination_path: target.pathname, link_text: target.textContent.trim().slice(0, 80) });
      } else if (target instanceof HTMLAnchorElement && target.hostname === "github.com") {
        track("github_link_clicked", { destination_path: target.pathname });
      }
    });

    const film = document.querySelector(".film video");
    film?.addEventListener("play", () => track("launch_film_started"), { once: true });
    film?.addEventListener("ended", () => track("launch_film_completed"), { once: true });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", setupAnalytics, { once: true });
  } else {
    setupAnalytics();
  }
})();
