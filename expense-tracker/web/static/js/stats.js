(function () {
    "use strict";

    var MONTHS = ["Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"];
    var COLORS = ["#4f46e5", "#16a34a", "#f59e0b", "#dc2626", "#0891b2",
                  "#9333ea", "#db2777", "#64748b", "#84cc16", "#ea580c"];
    var money = new Intl.NumberFormat("ru-RU", { style: "currency", currency: "RUB" });
    var charts = {};

    function $(id) { return document.getElementById(id); }

    function showError(id, message) {
        var el = $(id);
        el.textContent = message || "";
        el.hidden = !message;
    }

    // Запрос к JSON API. Cookie сессии отправляется автоматически (same-origin).
    async function fetchJSON(url) {
        var resp = await fetch(url, { headers: { "Accept": "application/json" }, credentials: "same-origin" });
        if (resp.status === 401) {
            window.location.href = "/login";
            throw new Error("требуется авторизация");
        }
        var data = await resp.json().catch(function () { return {}; });
        if (!resp.ok) {
            throw new Error(data.error || ("HTTP " + resp.status));
        }
        return data;
    }

    function drawChart(key, canvasId, config) {
        if (charts[key]) charts[key].destroy();
        charts[key] = new Chart($(canvasId), config);
    }

    async function loadByCategory() {
        var params = new URLSearchParams();
        var from = $("cat-from").value;
        var to = $("cat-to").value;
        if (from) params.set("from", from);
        if (to) params.set("to", to);

        try {
            var data = await fetchJSON("/api/stats/by-category?" + params.toString());
            showError("cat-error", "");
            $("cat-total").textContent = money.format(data.total);
            $("cat-empty").hidden = data.items.length > 0;

            drawChart("category", "chart-category", {
                type: "pie",
                data: {
                    labels: data.items.map(function (i) { return i.category; }),
                    datasets: [{
                        data: data.items.map(function (i) { return i.total; }),
                        backgroundColor: data.items.map(function (_, idx) { return COLORS[idx % COLORS.length]; })
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: { position: "bottom" },
                        tooltip: { callbacks: { label: function (ctx) { return ctx.label + ": " + money.format(ctx.parsed); } } }
                    }
                }
            });
        } catch (err) {
            showError("cat-error", "Не удалось загрузить статистику: " + err.message);
        }
    }

    async function loadByMonth() {
        var year = $("month-year").value;
        try {
            var data = await fetchJSON("/api/stats/by-month?year=" + encodeURIComponent(year));
            showError("month-error", "");
            $("month-total").textContent = money.format(data.total);

            drawChart("month", "chart-month", {
                type: "bar",
                data: {
                    labels: data.items.map(function (i) { return MONTHS[i.month - 1]; }),
                    datasets: [{
                        label: "Траты за " + data.year + " год",
                        data: data.items.map(function (i) { return i.total; }),
                        backgroundColor: "#4f46e5",
                        borderRadius: 6
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: { y: { beginAtZero: true, ticks: { callback: function (v) { return money.format(v); } } } },
                    plugins: {
                        legend: { display: false },
                        tooltip: { callbacks: { label: function (ctx) { return money.format(ctx.parsed.y); } } }
                    }
                }
            });
        } catch (err) {
            showError("month-error", "Не удалось загрузить статистику: " + err.message);
        }
    }

    if (typeof Chart === "undefined") {
        showError("cat-error", "Библиотека Chart.js не загрузилась (проверьте подключение к интернету).");
        showError("month-error", "Библиотека Chart.js не загрузилась (проверьте подключение к интернету).");
        return;
    }

    $("cat-apply").addEventListener("click", loadByCategory);
    $("month-year").addEventListener("change", loadByMonth);
    loadByCategory();
    loadByMonth();
})();
