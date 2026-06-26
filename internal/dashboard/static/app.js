const els = {
  filters: document.querySelector("#filters"),
  siteId: document.querySelector("#siteId"),
  grain: document.querySelector("#grain"),
  from: document.querySelector("#from"),
  to: document.querySelector("#to"),
  status: document.querySelector("#status"),
  rowCount: document.querySelector("#rowCount"),
  domainCount: document.querySelector("#domainCount"),
  pageCount: document.querySelector("#pageCount"),
  domainRows: document.querySelector("#domainRows"),
  pageRows: document.querySelector("#pageRows"),
  metrics: {
    page_views: document.querySelector("#metricPv"),
    ip_count: document.querySelector("#metricIp"),
    sessions: document.querySelector("#metricSessions"),
    active_visitors: document.querySelector("#metricActiveUv"),
    summed_active_visitors: document.querySelector("#metricSummedUv"),
    cumulative_visitors: document.querySelector("#metricCumulativeUv"),
    blended_visitors: document.querySelector("#metricBlendedUv")
  }
};

const chartBox = document.querySelector("#trendChart");
const chart = window.echarts ? echarts.init(chartBox) : null;
const fallbackCanvas = chart ? null : createFallbackCanvas(chartBox);
const numberFormat = new Intl.NumberFormat("zh-CN");

function qs(params) {
  const out = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value) out.set(key, value);
  });
  const text = out.toString();
  return text ? `?${text}` : "";
}

async function getJSON(path) {
  const response = await fetch(path, { headers: { Accept: "application/json" } });
  if (!response.ok) {
    throw new Error(`${response.status} ${response.statusText}`);
  }
  return response.json();
}

function value(row, key) {
  return row[key] ?? row[toPascal(key)] ?? 0;
}

function toPascal(key) {
  return key.split("_").map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join("");
}

function formatNumber(value) {
  return numberFormat.format(Number(value || 0));
}

function setStatus(text, isError = false) {
  els.status.textContent = text;
  els.status.classList.toggle("error", isError);
}

function setMetric(id, value) {
  els.metrics[id].textContent = formatNumber(value);
}

function normalizeOverview(overview, online) {
  return {
    page_views: overview.page_views || 0,
    ip_count: overview.ip_count || 0,
    sessions: overview.sessions || 0,
    active_visitors: overview.active_visitors ?? online.count ?? 0,
    summed_active_visitors: overview.summed_active_visitors ?? overview.unique_visitors ?? 0,
    cumulative_visitors: overview.cumulative_visitors ?? overview.unique_visitors ?? 0,
    blended_visitors: overview.blended_visitors ?? Math.max(overview.unique_visitors || 0, online.count || 0)
  };
}

function renderMetrics(overview, online) {
  const metrics = normalizeOverview(overview, online);
  Object.entries(metrics).forEach(([key, metricValue]) => setMetric(key, metricValue));
}

function renderChart(rows) {
 els.rowCount.textContent = `${rows.length} 个时间点`;
  const labels = rows.map((row) => new Date(value(row, "bucket")).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }));
  const runningUv = [];
  rows.reduce((sum, row) => {
    const next = sum + Number(value(row, "unique_visitors"));
    runningUv.push(next);
    return next;
  }, 0);

  const series = [
    ["PV", "page_views"],
    ["IP", "ip_count"],
    ["Session", "sessions"],
    ["UV", "unique_visitors"],
    ["活跃UV", "unique_visitors"],
    ["累计UV", null],
    ["融合UV", null]
  ].map(([name, key]) => ({
    name,
    type: "line",
    smooth: true,
    symbolSize: 5,
    data: key ? rows.map((row) => value(row, key)) : runningUv
  }));

  series[6].data = rows.map((row, index) => Math.max(Number(value(row, "unique_visitors")), runningUv[index] || 0));

  if (!chart) {
    drawFallbackChart(fallbackCanvas, labels, series);
    return;
  }

  chart.setOption({
    color: ["#2364aa", "#247a57", "#9b5de5", "#e07a5f", "#2a9d8f", "#f2b84b", "#b54747"],
    tooltip: { trigger: "axis" },
    legend: { top: 10, type: "scroll" },
    grid: { left: 48, right: 24, top: 58, bottom: 44 },
    xAxis: { type: "category", data: labels, boundaryGap: false },
    yAxis: { type: "value" },
    series
  });
}

function createFallbackCanvas(target) {
  const canvas = document.createElement("canvas");
  canvas.className = "fallbackCanvas";
  target.appendChild(canvas);
  return canvas;
}

function drawFallbackChart(canvas, labels, series) {
  if (!canvas) return;
  const rect = canvas.parentElement.getBoundingClientRect();
  const scale = window.devicePixelRatio || 1;
  canvas.width = Math.max(1, Math.round(rect.width * scale));
  canvas.height = Math.max(1, Math.round(rect.height * scale));
  canvas.style.width = `${Math.round(rect.width)}px`;
  canvas.style.height = `${Math.round(rect.height)}px`;
  const ctx = canvas.getContext("2d");
  ctx.scale(scale, scale);
  ctx.clearRect(0, 0, rect.width, rect.height);
  ctx.fillStyle = "#ffffff";
  ctx.fillRect(0, 0, rect.width, rect.height);

  const width = rect.width;
  const height = rect.height;
  const pad = { left: 50, right: 24, top: 34, bottom: 42 };
  const plotW = Math.max(1, width - pad.left - pad.right);
  const plotH = Math.max(1, height - pad.top - pad.bottom);
  const maxValue = Math.max(1, ...series.flatMap((item) => item.data.map((n) => Number(n) || 0)));
  const colors = ["#2364aa", "#247a57", "#9b5de5", "#e07a5f", "#2a9d8f", "#f2b84b", "#b54747"];

  ctx.strokeStyle = "#d8dee6";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(pad.left, pad.top);
  ctx.lineTo(pad.left, pad.top + plotH);
  ctx.lineTo(pad.left + plotW, pad.top + plotH);
  ctx.stroke();

  if (!labels.length) {
    ctx.fillStyle = "#667085";
    ctx.font = "14px system-ui";
    ctx.textAlign = "center";
    ctx.fillText("暂无趋势数据", width / 2, height / 2);
    return;
  }

  series.forEach((item, index) => {
    ctx.strokeStyle = colors[index % colors.length];
    ctx.lineWidth = 2;
    ctx.beginPath();
    item.data.forEach((raw, pointIndex) => {
      const x = pad.left + (labels.length === 1 ? 0 : (pointIndex / (labels.length - 1)) * plotW);
      const y = pad.top + plotH - ((Number(raw) || 0) / maxValue) * plotH;
      if (pointIndex === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    });
    ctx.stroke();
  });

  ctx.fillStyle = "#667085";
  ctx.font = "12px system-ui";
  ctx.textAlign = "left";
  ctx.fillText(labels[0] || "", pad.left, height - 16);
  ctx.textAlign = "right";
  ctx.fillText(labels[labels.length - 1] || "", width - pad.right, height - 16);
}

function referrerDomain(key) {
  if (!key) return "直接访问";
  try {
    return new URL(key).hostname || key;
  } catch {
    return key.split("/")[0] || key;
  }
}

function aggregateDomains(rows) {
  const map = new Map();
  rows.forEach((row) => {
    const domain = referrerDomain(value(row, "key"));
    const current = map.get(domain) || { key: domain, page_views: 0, ip_count: 0, unique_visitors: 0, sessions: 0 };
    current.page_views += Number(value(row, "page_views"));
    current.ip_count += Number(value(row, "ip_count"));
    current.unique_visitors += Number(value(row, "unique_visitors"));
    current.sessions += Number(value(row, "sessions"));
    map.set(domain, current);
  });
  return [...map.values()].sort((a, b) => b.page_views - a.page_views);
}

function renderTable(target, rows, emptyText) {
  target.innerHTML = "";
  if (!rows.length) {
    const tr = document.createElement("tr");
    tr.innerHTML = `<td class="empty" colspan="5">${emptyText}</td>`;
    target.appendChild(tr);
    return;
  }
  rows.slice(0, 20).forEach((row) => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td title="${escapeHtml(String(row.key || ""))}">${escapeHtml(String(row.key || ""))}</td>
      <td>${formatNumber(row.page_views)}</td>
      <td>${formatNumber(row.ip_count)}</td>
      <td>${formatNumber(row.unique_visitors)}</td>
      <td>${formatNumber(row.sessions)}</td>
    `;
    target.appendChild(tr);
  });
}

function escapeHtml(text) {
  return text.replace(/[&<>"']/g, (char) => ({
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#039;"
  })[char]);
}

function normalizeDimensionRows(rows) {
  return rows.map((row) => ({
    key: value(row, "key"),
    page_views: Number(value(row, "page_views")),
    ip_count: Number(value(row, "ip_count")),
    unique_visitors: Number(value(row, "unique_visitors")),
    sessions: Number(value(row, "sessions")),
    event_count: Number(value(row, "event_count"))
  }));
}

async function loadDashboard() {
  const siteID = els.siteId.value.trim();
  if (!siteID) {
    setStatus("请输入站点 ID", true);
    return;
  }
  const params = qs({ grain: els.grain.value, from: els.from.value, to: els.to.value });
  const dimensionParams = qs({ dimension: "referrer", from: els.from.value, to: els.to.value, limit: "100" });
  setStatus("加载中...");
  try {
    const base = `/api/sites/${encodeURIComponent(siteID)}`;
    const [online, overview, trend, referrers] = await Promise.all([
      getJSON(`${base}/online`),
      getJSON(`${base}/overview${params}`),
      getJSON(`${base}/trend${params}`),
      getJSON(`${base}/dimensions${dimensionParams}`)
    ]);
    const trendRows = trend.rows || trend.Rows || [];
    const referrerRows = normalizeDimensionRows(referrers.rows || referrers.Rows || []);
    renderMetrics(overview, online);
    renderChart(trendRows);
    const domainRows = aggregateDomains(referrerRows);
    els.domainCount.textContent = `${domainRows.length} 条`;
    els.pageCount.textContent = `${referrerRows.length} 条`;
    renderTable(els.domainRows, domainRows, "暂无来路域名");
    renderTable(els.pageRows, referrerRows, "暂无来路页面");
    setStatus(`已更新 ${new Date().toLocaleTimeString("zh-CN")}`);
  } catch (error) {
    setStatus(`加载失败: ${error.message}`, true);
  }
}

els.filters.addEventListener("submit", (event) => {
  event.preventDefault();
  loadDashboard();
});

window.addEventListener("resize", () => {
  if (chart) chart.resize();
  else loadDashboard();
});

loadDashboard();
