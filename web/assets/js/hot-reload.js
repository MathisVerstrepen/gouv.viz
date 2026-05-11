(function () {
	"use strict";

	const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
	const socket = new WebSocket(protocol + "//" + window.location.host + "/ws");

	window.addEventListener("beforeunload", function () {
		socket.close();
	});

	socket.addEventListener("close", function (event) {
		if (event.code === 1000) {
			return;
		}

		const interval = window.setInterval(function () {
			window.fetch("/ping").then(function (response) {
				if (!response.ok) {
					return;
				}

				window.clearInterval(interval);
				window.setTimeout(function () {
					window.location.reload();
				}, 100);
			});
		}, 200);
	});
})();
