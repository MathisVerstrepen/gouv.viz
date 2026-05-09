(function () {
	"use strict";

	const storageKey = "gouvviz.theme";
	const root = document.documentElement;
	const media = window.matchMedia("(prefers-color-scheme: dark)");

	function storedScheme() {
		try {
			const value = window.localStorage.getItem(storageKey);
			return value === "light" || value === "dark" ? value : "system";
		} catch (_error) {
			return "system";
		}
	}

	function currentScheme() {
		const scheme = root.getAttribute("data-fr-scheme") || "system";
		return scheme === "light" || scheme === "dark" ? scheme : "system";
	}

	function resolvedScheme() {
		const scheme = currentScheme();
		if (scheme !== "system") {
			return scheme;
		}
		return media.matches ? "dark" : "light";
	}

	function applyScheme(scheme) {
		if (scheme === "light" || scheme === "dark") {
			root.setAttribute("data-fr-scheme", scheme);
			return;
		}

		root.setAttribute("data-fr-scheme", "system");
	}

	function persistScheme(scheme) {
		try {
			window.localStorage.setItem(storageKey, scheme);
		} catch (_error) {
			// Theme persistence is optional; the button still works for this page.
		}
	}

	function updateToggle(button) {
		const label = button.querySelector("[data-theme-toggle-label]");
		const isDark = resolvedScheme() === "dark";
		const nextLabel = isDark ? "Mode clair" : "Mode sombre";

		button.setAttribute("aria-pressed", String(!isDark));
		button.setAttribute("aria-label", "Activer le " + nextLabel.toLowerCase());
		if (label) {
			label.textContent = nextLabel;
		}
	}

	function setupToggle() {
		const button = document.querySelector("[data-theme-toggle]");
		if (!button) {
			return;
		}

		updateToggle(button);
		button.addEventListener("click", function () {
			const nextScheme = resolvedScheme() === "dark" ? "light" : "dark";
			applyScheme(nextScheme);
			persistScheme(nextScheme);
			updateToggle(button);
		});

		media.addEventListener("change", function () {
			if (currentScheme() === "system") {
				updateToggle(button);
			}
		});
	}

	applyScheme(storedScheme());

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", setupToggle);
	} else {
		setupToggle();
	}
})();
