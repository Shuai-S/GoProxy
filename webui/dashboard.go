package webui

const dashboardHTML = `<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>GoProxy — 智能代理池</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700&family=Share+Tech+Mono&display=swap" rel="stylesheet">
<style>
*{box-sizing:border-box;margin:0;padding:0}
:root{
  --bg:#0a0a0a;
  --bg-elevated:#111;
  --bg-card:#0d0d0d;
  --fg:#00ff41;
  --fg-dim:#00cc33;
  --fg-text:#0f0;
  --border:#1a3a1a;
  --border-heavy:#00ff41;
  --gray-1:#0d0d0d;
  --gray-2:#151515;
  --gray-3:#1a1a1a;
  --gray-4:#2a4a2a;
  --gray-5:#00aa2a;
  --gray-6:#00dd38;
  --green:#00ff41;
  --yellow:#ffff00;
  --orange:#ff8800;
  --red:#ff0033;
  --mono:JetBrains Mono,Share Tech Mono,monospace;
  --sans:JetBrains Mono,monospace;
}
body{background:var(--bg);color:var(--fg);font-family:var(--mono);font-size:14px;line-height:1.5;-webkit-font-smoothing:antialiased;position:relative}

/* CRT 扫描线效果 */
body::before{content:'';position:fixed;top:0;left:0;width:100%;height:100%;background:repeating-linear-gradient(0deg,rgba(0,255,65,0.03) 0px,transparent 1px,transparent 2px,rgba(0,255,65,0.03) 3px);pointer-events:none;z-index:9999}

/* 荧光光晕效果 */
body::after{content:'';position:fixed;top:0;left:0;width:100%;height:100%;background:radial-gradient(ellipse at center,rgba(0,255,65,0.05) 0%,transparent 70%);pointer-events:none;z-index:9998}

.layout{max-width:1800px;margin:0 auto;padding:0 32px}

/* 双列布局 */
.content-grid{display:grid;grid-template-columns:1fr 420px;gap:32px;align-items:start}
.main-content{min-width:0;position:relative}
.sidebar{position:sticky;top:32px}

/* 控制面板 */
.control-panel{background:var(--bg-card);border:1px solid var(--border-heavy);padding:20px;margin-bottom:20px;box-shadow:0 0 20px rgba(0,255,65,0.15)}
.control-header{display:flex;align-items:center;justify-content:center;margin-bottom:16px;padding-bottom:12px;border-bottom:1px solid var(--border)}
.control-title{font-size:14px;font-weight:700;letter-spacing:0.12em;font-family:var(--mono);text-transform:uppercase;color:var(--fg);text-shadow:0 0 10px var(--fg)}
.control-ops{display:flex;flex-direction:row;gap:8px}
.ctrl-btn-primary{width:100%;padding:10px;font-size:10px;font-weight:600;cursor:pointer;border:1px solid var(--border-heavy);background:var(--bg-card);color:var(--fg);font-family:var(--mono);text-transform:uppercase;letter-spacing:0.08em;transition:all 0.2s}
.ctrl-btn-primary:hover{background:var(--border);box-shadow:0 0 15px var(--border-heavy);color:var(--fg);text-shadow:0 0 5px var(--fg)}
.ctrl-btn-secondary{width:100%;padding:8px;font-size:9px;font-weight:600;cursor:pointer;border:1px solid var(--border);background:var(--bg-card);color:var(--fg-dim);font-family:var(--mono);text-transform:uppercase;letter-spacing:0.08em;transition:all 0.2s}
.ctrl-btn-secondary:hover{background:var(--border);color:var(--fg);box-shadow:0 0 8px var(--border)}
.proxy-config{margin-top:14px;padding-top:14px;border-top:1px solid var(--border)}
.proxy-config-title{font-size:8px;text-transform:uppercase;letter-spacing:0.14em;color:var(--fg-dim);margin-bottom:8px;font-weight:600}
.proxy-config-list{display:flex;flex-direction:column;gap:6px}
.proxy-config-row{display:grid;grid-template-columns:88px 1fr 34px;align-items:center;gap:6px;border:1px solid var(--border);background:var(--bg);padding:6px}
.proxy-config-label{font-size:8px;color:var(--fg-dim);letter-spacing:0.08em;text-transform:uppercase}
.proxy-config-address{min-width:0;font-size:9px;color:var(--fg);font-family:var(--mono);white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.proxy-copy-btn{height:26px;border:1px solid var(--border);background:var(--bg-card);color:var(--fg-dim);cursor:pointer;font-size:10px;font-family:var(--mono);transition:all 0.2s}
.proxy-copy-btn:hover{background:var(--border);color:var(--fg);box-shadow:0 0 8px var(--border)}

/* 代理列表区域 */
.proxy-section{display:block}
.proxy-header{position:sticky;top:0;z-index:100;background:var(--bg);padding:20px 0 16px;border-bottom:1px solid var(--border-heavy);display:flex;align-items:center;justify-content:space-between;gap:24px;backdrop-filter:blur(8px);box-shadow:0 2px 0 0 rgba(0,255,65,0.2)}
.proxy-logo-area{display:flex;align-items:baseline;gap:12px;flex-shrink:0}
.proxy-logo{font-size:28px;font-weight:900;letter-spacing:0.2em;font-family:var(--mono);text-transform:uppercase;color:var(--fg);text-shadow:0 0 15px var(--fg),0 0 30px var(--fg);animation:glow 2s ease-in-out infinite alternate}
@keyframes glow{0%{text-shadow:0 0 15px var(--fg),0 0 30px var(--fg)}100%{text-shadow:0 0 20px var(--fg),0 0 40px var(--fg),0 0 60px var(--fg)}}
.user-badge{font-size:10px;color:var(--fg-dim);font-family:var(--mono);letter-spacing:0.08em;opacity:0.6}
.proxy-content{overflow-x:auto}
.header-actions{display:flex;gap:8px;align-items:center;flex-shrink:0}

/* 响应式：屏幕小于1200px时变为单列 */
@media (max-width: 1200px) {
  .content-grid{grid-template-columns:1fr;height:auto}
  .sidebar{overflow-y:visible;padding-right:0}
  .main-content{overflow:visible}
  .proxy-section{height:auto;overflow:visible}
  .proxy-content{overflow-y:visible}
}

/* Health Grid - 侧边栏紧凑布局 */
.health-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:1px;background:var(--bg);border:1px solid var(--border);margin-bottom:10px;box-shadow:0 0 20px rgba(0,255,65,0.1)}
.health-card{background:var(--bg-card);padding:8px 10px;position:relative;border:1px solid var(--border)}
.health-label{font-size:8px;text-transform:uppercase;letter-spacing:0.15em;color:var(--fg-dim);margin-bottom:4px;font-weight:600;font-family:var(--mono)}
.health-value{font-size:18px;font-weight:700;font-family:var(--mono);line-height:1;letter-spacing:0.05em;color:var(--fg);text-shadow:0 0 10px var(--fg)}
.health-status{position:absolute;top:16px;right:16px;width:8px;height:8px;border-radius:50%}
.health-status.healthy{background:var(--green);box-shadow:0 0 8px var(--green)}
.health-status.warning{background:var(--orange);box-shadow:0 0 8px var(--orange)}
.health-status.critical{background:var(--red);box-shadow:0 0 8px var(--red)}
.health-status.emergency{background:var(--red);box-shadow:0 0 15px var(--red),0 0 0 3px rgba(255,0,51,0.3);animation:pulse 1s infinite}
.health-meta{font-size:8px;color:var(--gray-5);margin-top:3px;font-family:var(--mono)}

@keyframes pulse{0%,100%{opacity:1}50%{opacity:0.6}}

/* Tabs/按钮样式 */
.tab{padding:8px 16px;min-height:36px;font-size:10px;font-weight:600;cursor:pointer;border:1px solid var(--border);background:var(--bg-card);color:var(--fg-dim);font-family:var(--mono);transition:all 0.2s;text-transform:uppercase;letter-spacing:0.05em;display:inline-flex;align-items:center;justify-content:center;text-decoration:none;box-sizing:border-box}
.tab:hover{background:var(--border);color:var(--fg);box-shadow:0 0 8px var(--border)}

/* 筛选下拉框 */
.filter-select{padding:8px 16px;min-height:36px;font-size:10px;font-weight:600;cursor:pointer;border:1px solid var(--border);background:var(--bg-card);color:var(--fg-dim);font-family:var(--mono);text-transform:uppercase;letter-spacing:0.05em;transition:all 0.2s;outline:none;appearance:none;background-image:url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%2300ff41' d='M6 9L1 4h10z'/%3E%3C/svg%3E");background-repeat:no-repeat;background-position:right 8px center;padding-right:32px}
.filter-select:hover{background-color:var(--border);color:var(--fg);box-shadow:0 0 8px var(--border)}
.filter-select option{background:var(--bg-card);color:var(--fg-dim)}

/* Quality Bar - 侧边栏紧凑布局 */
.quality-bar{background:var(--bg-card);border:1px solid var(--border);padding:16px;margin-bottom:16px;box-shadow:0 0 15px rgba(0,255,65,0.08)}
.quality-bar-title{font-size:8px;text-transform:uppercase;letter-spacing:0.15em;color:var(--fg-dim);margin-bottom:10px;font-weight:600}
.quality-visual{display:flex;height:20px;border:1px solid var(--border);overflow:hidden;box-shadow:inset 0 0 10px rgba(0,255,65,0.1)}
.quality-segment{display:flex;align-items:center;justify-content:center;font-size:9px;font-weight:700;font-family:var(--mono);color:#000;transition:width 0.3s;text-shadow:none}
.quality-s{background:var(--green);box-shadow:0 0 10px var(--green)}
.quality-a{background:var(--yellow);box-shadow:0 0 10px var(--yellow)}
.quality-b{background:var(--orange);box-shadow:0 0 10px var(--orange)}
.quality-c{background:var(--red);box-shadow:0 0 10px var(--red)}
.quality-legend{display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-top:10px}
.quality-legend-item{font-size:9px;font-family:var(--mono);color:var(--fg-dim)}
.quality-legend-dot{display:inline-block;width:6px;height:6px;margin-right:5px;box-shadow:0 0 4px currentColor}

/* 操作按钮样式 */
.btn-danger{border:1px solid var(--red);color:var(--red);padding:5px 10px;font-size:9px;text-transform:uppercase;letter-spacing:0.08em;background:var(--bg-card);cursor:pointer;transition:all 0.2s}
.btn-danger:hover{background:var(--red);color:#000;box-shadow:0 0 10px var(--red)}
.btn-action{border:1px solid var(--border);color:var(--fg-dim);padding:5px 10px;font-size:9px;text-transform:uppercase;letter-spacing:0.08em;background:var(--bg-card);cursor:pointer;transition:all 0.2s}
.btn-action:hover{background:var(--border);color:var(--fg);box-shadow:0 0 8px var(--border)}

/* Table */
table{width:100%;min-width:1080px;border-collapse:collapse;font-size:11px;font-family:var(--mono);border:1px solid var(--border);background:var(--bg-card);table-layout:fixed}
thead{position:relative;z-index:1;border-bottom:1px solid var(--border-heavy);background:var(--bg-elevated);box-shadow:0 2px 8px rgba(0,0,0,0.3)}
thead tr{height:54px}
th{height:54px;padding:8px 12px;text-align:left;font-size:9px;text-transform:uppercase;letter-spacing:0.08em;color:var(--fg-dim);font-weight:600;line-height:1.35;vertical-align:middle;white-space:normal;word-break:keep-all;overflow-wrap:normal}
td{height:78px;padding:12px;border-bottom:1px solid var(--border);color:var(--fg-dim);vertical-align:middle}
tr:last-child td{border-bottom:none}
tr:hover{background:var(--gray-2);box-shadow:inset 0 0 20px rgba(0,255,65,0.05)}
.cell-mono{font-family:var(--mono);font-size:10px}
.cell-grade{font-weight:700;font-size:14px}
.cell-stack{display:flex;flex-direction:column;gap:4px;min-width:0}
.cell-main{min-width:0;overflow:hidden;text-overflow:ellipsis}
.cell-sub{font-size:9px;color:var(--gray-5);line-height:1.2;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.cell-protocol{display:flex;flex-direction:column;align-items:flex-start;gap:4px}
.subscription-badge{display:inline-block;background:var(--yellow);color:#000;font-size:8px;font-weight:700;padding:1px 5px;letter-spacing:0.05em;max-width:100%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.cell-action{width:80px}
.action-stack{display:flex;flex-direction:column;align-items:stretch;gap:6px;width:58px;margin:0 auto}
.action-stack button{width:58px;height:30px;padding:0;text-align:center}
.th-grade{width:48px}
.th-protocol{width:138px}
.th-address{width:180px}
.th-exit-ip{width:150px}
.th-location{width:220px}
.th-risk{width:128px}
.th-latency{width:78px}
.th-usage{width:84px}
.th-action{width:86px}
.cell-clickable{cursor:pointer;transition:all 0.2s}
.cell-clickable:hover{background:var(--border)!important;color:var(--fg)!important;box-shadow:0 0 8px var(--border)!important}
.cell-clickable:active{background:var(--border-heavy)!important}
.grade-s{color:var(--green);text-shadow:0 0 8px var(--green)}
.grade-a{color:var(--yellow);text-shadow:0 0 8px var(--yellow)}
.grade-b{color:var(--orange);text-shadow:0 0 8px var(--orange)}
.grade-c{color:var(--red);text-shadow:0 0 8px var(--red)}
.badge{display:inline-block;padding:3px 8px;font-size:9px;font-weight:600;text-transform:uppercase;letter-spacing:0.08em;border:1px solid;font-family:var(--mono)}
.badge-http{border-color:var(--fg-dim);color:var(--fg-dim);background:transparent}
.badge-socks5{background:var(--fg-dim);color:#000;border-color:var(--fg-dim);box-shadow:0 0 6px var(--fg-dim)}
.badge-risk-low{border-color:var(--green);color:var(--green);background:rgba(0,255,65,0.08)}
.badge-risk-med{border-color:var(--yellow);color:var(--yellow);background:rgba(255,255,0,0.08)}
.badge-risk-high{border-color:var(--red);color:var(--red);background:rgba(255,0,51,0.08)}
.badge-muted{border-color:var(--border);color:var(--fg-dim);background:transparent}
.latency{font-weight:600}
.latency-excellent{color:var(--green)}
.latency-good{color:#333}
.latency-fair{color:#666}
.latency-poor{color:var(--red)}

/* Modal */
.modal-overlay{display:none;position:fixed;inset:0;background:rgba(0,0,0,0.95);backdrop-filter:blur(10px);z-index:100;align-items:center;justify-content:center}
.modal-overlay.show{display:flex}
.modal{background:var(--bg-elevated);border:1px solid var(--border-heavy);padding:40px;width:700px;box-shadow:0 0 40px rgba(0,255,65,0.3);max-height:90vh;overflow-y:auto}
.modal-title{font-size:20px;font-weight:700;margin-bottom:28px;letter-spacing:0.08em;text-transform:uppercase;color:var(--fg);text-shadow:0 0 10px var(--fg)}
.form-section{margin-bottom:28px}
.form-section-title{font-size:9px;text-transform:uppercase;letter-spacing:0.12em;color:var(--fg-dim);margin-bottom:12px;font-weight:600;padding-bottom:8px;border-bottom:1px solid var(--border)}
.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.form-group{display:flex;flex-direction:column}
.form-group label{font-size:9px;text-transform:uppercase;letter-spacing:0.08em;color:var(--fg-dim);margin-bottom:6px;font-weight:600}
.form-group input{padding:10px;background:var(--bg-card);border:1px solid var(--border);font-size:12px;font-family:var(--mono);color:var(--fg);outline:none;transition:all 0.2s}
.form-group input:focus{border-color:var(--border-heavy);background:var(--bg-elevated);box-shadow:0 0 10px var(--border-heavy)}
.form-help{font-size:9px;color:var(--gray-5);margin-top:4px;font-family:var(--mono)}
.modal-actions{display:flex;gap:12px;margin-top:28px;padding-top:28px;border-top:1px solid var(--border)}
.modal-actions .btn{flex:1;padding:12px 24px;font-size:11px;font-weight:600;cursor:pointer;border:1px solid var(--border-heavy);background:var(--bg-card);color:var(--fg);font-family:var(--mono);text-transform:uppercase;letter-spacing:0.08em;transition:all 0.2s}
.modal-actions .btn:hover{background:var(--border);box-shadow:0 0 15px var(--border-heavy);color:var(--fg);text-shadow:0 0 5px var(--fg)}
.modal-actions .btn-secondary{border:1px solid var(--border);background:var(--bg-card);color:var(--fg-dim)}
.modal-actions .btn-secondary:hover{background:var(--gray-2);color:var(--fg);box-shadow:0 0 10px var(--border)}

/* Log - 适配侧边栏布局 */
.log-box{padding:12px;background:var(--bg);border:1px solid var(--border);font-family:var(--mono);font-size:10px;color:var(--fg-dim);height:350px;overflow-y:auto;line-height:1.8;box-shadow:inset 0 0 20px rgba(0,255,65,0.05)}
.log-box::-webkit-scrollbar{width:4px}
.log-box::-webkit-scrollbar-track{background:var(--bg)}
.log-box::-webkit-scrollbar-thumb{background:var(--border);border-radius:2px}
.log-box::-webkit-scrollbar-thumb:hover{background:var(--border-heavy)}
.log-line{padding:3px 0;opacity:0.85}
.log-line.error{color:var(--red);font-weight:600;text-shadow:0 0 5px var(--red)}
.log-line.success{color:var(--green);text-shadow:0 0 5px var(--green)}

/* 侧边栏样式 */
.sidebar>*:not(:last-child){margin-bottom:16px}
.sidebar .section{margin-bottom:0;border:1px solid var(--border);background:var(--bg-card);padding:16px;box-shadow:0 0 15px rgba(0,255,65,0.1)}
.sidebar .section-header{padding-bottom:10px;margin-bottom:12px;border-bottom:1px solid var(--border)}
.sidebar .section-title{font-size:12px;letter-spacing:0.12em}

/* 响应式布局 */
@media (max-width: 1200px) {
  .content-grid{grid-template-columns:1fr}
  .sidebar{position:static}
  .health-grid{grid-template-columns:repeat(4,1fr)}
  .health-card{padding:10px 12px}
  .health-value{font-size:32px}
  .log-box{height:400px}
  .sidebar .section{border:1px solid var(--border)}
}

.empty{padding:48px;text-align:center;color:var(--gray-4);font-size:12px;font-family:var(--mono);text-transform:uppercase;letter-spacing:0.08em}

/* 权限控制 - 默认隐藏管理员功能 */
.admin-only{display:none}

/* Toast 提示 */
.toast{position:fixed;bottom:32px;left:50%;transform:translateX(-50%) translateY(100px);background:var(--fg);color:#000;padding:12px 24px;font-size:11px;font-weight:600;font-family:var(--mono);opacity:0;transition:all 0.3s;z-index:1000;pointer-events:none;box-shadow:0 0 20px var(--fg);text-transform:uppercase;letter-spacing:0.05em}
.toast.show{transform:translateX(-50%) translateY(0);opacity:1}
</style>
</head>
<body>
<div class="layout">
  <div class="content-grid">
    <div class="main-content">
      <div class="proxy-section">
        <div class="proxy-header">
          <div class="proxy-logo-area">
            <div class="proxy-logo">[ GoProxy ]</div>
            <span id="user-mode" class="user-badge">admin</span>
          </div>
          <div class="header-actions">
            <select class="filter-select" id="protocol-filter" onchange="setProtocolFilter(this.value)">
              <option value="" id="protocol-filter-label">协议</option>
              <option value="http">HTTP</option>
              <option value="socks5">SOCKS5</option>
            </select>
            <select class="filter-select" id="country-filter" onchange="setCountryFilter(this.value)">
              <option value="" id="country-filter-label">出口国家</option>
            </select>
            <button class="tab" onclick="toggleLang()" id="lang-btn">[ EN ]</button>
            <a href="https://github.com/isboyjc/ProxyGo" target="_blank" class="tab" title="GitHub">
              <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor" style="vertical-align: middle;">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
              </svg>
            </a>
            <button class="tab admin-only" onclick="openContributeModal()" style="color:var(--yellow)" data-i18n="contribute.nav">贡献订阅</button>
            <a href="/login" class="tab" id="login-link" style="display: none;" data-i18n="nav.login">登录</a>
            <a href="/logout" class="tab admin-only" data-i18n="nav.logout">退出</a>
            <button class="tab admin-only" onclick="openSettings()" title="" data-i18n-title="contribute.settings" style="padding:4px 8px">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
            </button>
          </div>
        </div>
        <div class="proxy-content">
          <div id="proxy-table-wrap"><div class="empty" data-i18n="proxy.loading">加载中...</div></div>
        </div>
      </div>
    </div>

    <aside class="sidebar">
      <div class="control-panel admin-only">
        <div class="control-header">
          <div class="control-title">[ CONTROL_PANEL ]</div>
        </div>
        <div class="control-ops">
          <button class="ctrl-btn-primary" onclick="triggerFetch()" data-i18n="actions.fetch">抓取代理</button>
          <button class="ctrl-btn-secondary" onclick="refreshLatency()" data-i18n="actions.refresh">刷新延迟</button>
          <button class="ctrl-btn-secondary" onclick="openManualModal()" data-i18n="proxy.add_manual">新增节点</button>
          <!-- 配置按钮已移到顶部导航 -->
        </div>
        <div class="proxy-config">
          <div class="proxy-config-title" data-i18n="control.proxy_config">代理配置地址</div>
          <div class="proxy-config-list" id="proxy-config-list">
            <div class="proxy-config-row">
              <span class="proxy-config-label">HTTP R</span>
              <span class="proxy-config-address">—</span>
              <button class="proxy-copy-btn" type="button">⧉</button>
            </div>
          </div>
        </div>
      </div>

      <!-- 订阅管理面板 -->
      <div class="control-panel admin-only" style="margin-bottom:20px">
        <div class="control-header">
          <div class="control-title">[ SUBSCRIPTIONS ]</div>
        </div>
        <div id="sub-list" style="margin-bottom:8px;font-size:11px;max-height:200px;overflow-y:auto"></div>
        <div class="control-ops">
          <button class="ctrl-btn-primary" onclick="openSubModal()" data-i18n="sub.add">添加订阅</button>
          <button class="ctrl-btn-secondary" onclick="refreshAllSubs()" data-i18n="sub.refresh_all">刷新所有订阅</button>
        </div>
        <div id="sub-status" style="margin-top:8px;font-size:10px;color:var(--fg-dim)"></div>
      </div>

      <!-- 免费代理池 -->
      <div style="font-size:8px;color:var(--fg-dim);letter-spacing:0.1em;text-transform:uppercase;margin-bottom:2px;font-weight:600" data-i18n="health.free_pool">[ FREE_POOL ]</div>
      <div class="health-grid">
        <div class="health-card">
          <div class="health-label" data-i18n="health.status">池子状态</div>
          <div class="health-value" id="pool-state" style="font-size:18px;text-transform:uppercase">—</div>
          <div class="health-status" id="pool-status-dot"></div>
        </div>
        <div class="health-card">
          <div class="health-label" data-i18n="health.total">免费代理</div>
          <div class="health-value" id="stat-total">0</div>
          <div class="health-meta"><span id="stat-capacity">0</span> <span data-i18n="health.capacity">容量</span></div>
        </div>
        <div class="health-card">
          <div class="health-label">HTTP</div>
          <div class="health-value" id="stat-http">0</div>
          <div class="health-meta"><span id="http-slots">0</span> <span data-i18n="health.slots">槽位</span> · <span id="http-avg">—</span>ms <span data-i18n="health.avg">平均</span></div>
        </div>
        <div class="health-card">
          <div class="health-label">SOCKS5</div>
          <div class="health-value" id="stat-socks5">0</div>
          <div class="health-meta"><span id="socks5-slots">0</span> <span data-i18n="health.slots">槽位</span> · <span id="socks5-avg">—</span>ms <span data-i18n="health.avg">平均</span></div>
        </div>
      </div>

      <!-- 订阅代理池 -->
      <div style="font-size:8px;color:var(--yellow);letter-spacing:0.1em;text-transform:uppercase;margin-bottom:2px;font-weight:600" data-i18n="health.sub_pool">[ SUBSCRIPTION_POOL ]</div>
      <div class="health-grid" style="grid-template-columns:repeat(3,1fr)">
        <div class="health-card">
          <div class="health-label" data-i18n="health.sub_sources">订阅源</div>
          <div class="health-value" id="stat-sub-count">0</div>
          <div class="health-meta" id="stat-sub-meta">—</div>
        </div>
        <div class="health-card">
          <div class="health-label" data-i18n="health.available">可用</div>
          <div class="health-value" id="stat-custom">0</div>
          <div class="health-meta" id="custom-meta">—</div>
        </div>
        <div class="health-card">
          <div class="health-label" data-i18n="health.disabled">禁用/待恢复</div>
          <div class="health-value" id="stat-custom-disabled">0</div>
          <div class="health-meta" id="custom-disabled-meta" data-i18n="health.awaiting_probe">探测唤醒中</div>
        </div>
      </div>

      <div class="quality-bar">
        <div class="quality-bar-title" data-i18n="quality.title">质量分布</div>
        <div class="quality-visual" id="quality-visual">
          <div class="quality-segment quality-s" style="width:0%"></div>
          <div class="quality-segment quality-a" style="width:0%"></div>
          <div class="quality-segment quality-b" style="width:0%"></div>
          <div class="quality-segment quality-c" style="width:0%"></div>
        </div>
        <div class="quality-legend">
          <div class="quality-legend-item"><span class="quality-legend-dot" style="background:#22c55e"></span><span data-i18n="quality.grade_s">S级</span> (<span id="grade-s-count">0</span>)</div>
          <div class="quality-legend-item"><span class="quality-legend-dot" style="background:#eab308"></span><span data-i18n="quality.grade_a">A级</span> (<span id="grade-a-count">0</span>)</div>
          <div class="quality-legend-item"><span class="quality-legend-dot" style="background:#f97316"></span><span data-i18n="quality.grade_b">B级</span> (<span id="grade-b-count">0</span>)</div>
          <div class="quality-legend-item"><span class="quality-legend-dot" style="background:#ef4444"></span><span data-i18n="quality.grade_c">C级</span> (<span id="grade-c-count">0</span>)</div>
        </div>
      </div>

      <div class="section">
        <div class="section-header">
          <h2 class="section-title" data-i18n="log.title">系统日志</h2>
        </div>
        <div class="log-box" id="logs-box"><span data-i18n="log.loading">加载中...</span></div>
        <div style="font-size:10px;color:var(--gray-5);font-family:var(--mono);margin-top:8px;text-align:center">
          <span data-i18n="log.auto_refresh_label">自动刷新</span>: <span id="log-countdown" style="color:var(--fg-dim);font-weight:600">5</span>s
        </div>
      </div>
    </aside>
  </div>
</div>

<div class="modal-overlay" id="settings-modal" onclick="if(event.target===this) closeSettings()">
  <div class="modal">
    <div class="modal-title" data-i18n="config.system_title">系统设置</div>

    <div class="form-section">
      <div class="form-section-title" data-i18n="config.section_proxy_mode">代理使用模式</div>
      <div class="form-grid">
        <div class="form-group" style="grid-column:1/-1">
          <label data-i18n="config.proxy_strategy">出站代理选择策略</label>
          <select id="cfg-custom-mode" style="width:100%;padding:10px;background:var(--bg-card);border:1px solid var(--border);color:var(--fg);font-family:var(--mono);font-size:12px">
            <option value="mixed_custom_priority" data-i18n="config.mode_mixed_custom">混合 · 订阅优先</option>
            <option value="mixed_free_priority" data-i18n="config.mode_mixed_free">混合 · 免费优先</option>
            <option value="mixed" data-i18n="config.mode_mixed">混合 · 平等</option>
            <option value="custom_only" data-i18n="config.mode_custom_only">仅订阅代理</option>
            <option value="free_only" data-i18n="config.mode_free_only">仅免费代理</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 免费池设置 -->
    <div class="form-section">
      <div class="form-section-title" data-i18n="config.section_free_pool">免费代理池</div>
      <div class="form-grid">
        <div class="form-group">
          <label data-i18n="config.pool_capacity">池子容量</label>
          <input type="number" id="cfg-pool-size" min="10" max="500">
          <div class="form-help" data-i18n="config.pool_capacity_help">免费代理总槽位</div>
        </div>
        <div class="form-group">
          <label data-i18n="config.http_ratio_label">HTTP 占比</label>
          <input type="number" id="cfg-http-ratio" min="0" max="1" step="0.05">
          <div class="form-help" data-i18n="config.http_ratio_help">0.3 = 30% HTTP</div>
        </div>
        <div class="form-group">
          <label data-i18n="config.min_per_protocol">每协议最小数</label>
          <input type="number" id="cfg-min-per-protocol" min="1" max="50">
        </div>
        <div class="form-group">
          <label data-i18n="config.latency_standard">标准延迟 (ms)</label>
          <input type="number" id="cfg-max-latency" min="500" max="5000" step="100">
        </div>
        <div class="form-group">
          <label data-i18n="config.latency_healthy">健康延迟 (ms)</label>
          <input type="number" id="cfg-max-latency-healthy" min="500" max="3000" step="100">
        </div>
        <div class="form-group">
          <label data-i18n="config.latency_emergency">紧急延迟 (ms)</label>
          <input type="number" id="cfg-max-latency-emergency" min="1000" max="5000" step="100">
        </div>
        <div class="form-group">
          <label data-i18n="config.optimize_interval">优化间隔 (分钟)</label>
          <input type="number" id="cfg-optimize-interval" min="10" max="120" step="10">
        </div>
        <div class="form-group">
          <label data-i18n="config.replace_threshold">替换阈值</label>
          <input type="number" id="cfg-replace-threshold" min="0.5" max="0.9" step="0.05">
          <div class="form-help" data-i18n="config.replace_threshold_help">新代理需快30%</div>
        </div>
      </div>
    </div>

    <!-- 订阅池设置 -->
    <div class="form-section">
      <div class="form-section-title" data-i18n="config.section_sub_pool">订阅代理池</div>
      <div class="form-grid">
        <div class="form-group">
          <label data-i18n="config.probe_interval">探测间隔 (分钟)</label>
          <input type="number" id="cfg-custom-probe" min="5" max="120" step="5">
          <div class="form-help" data-i18n="config.probe_interval_help">禁用代理的唤醒探测间隔</div>
        </div>
        <div class="form-group">
          <label data-i18n="config.refresh_interval">默认刷新间隔 (分钟)</label>
          <input type="number" id="cfg-custom-refresh" min="10" max="1440" step="10">
          <div class="form-help" data-i18n="config.refresh_interval_help">新订阅的默认刷新周期</div>
        </div>
      </div>
    </div>

    <!-- 验证与检查 -->
    <div class="form-section">
      <div class="form-section-title" data-i18n="config.section_validation">验证与健康检查</div>
      <div class="form-grid">
        <div class="form-group">
          <label data-i18n="config.validate_concurrency">验证并发数</label>
          <input type="number" id="cfg-concurrency" min="50" max="500" step="50">
        </div>
        <div class="form-group">
          <label data-i18n="config.validate_timeout">验证超时 (秒)</label>
          <input type="number" id="cfg-timeout" min="3" max="15">
        </div>
        <div class="form-group">
          <label data-i18n="config.health_interval">检查间隔 (分钟)</label>
          <input type="number" id="cfg-health-interval" min="1" max="60">
        </div>
        <div class="form-group">
          <label data-i18n="config.health_batch">每批数量</label>
          <input type="number" id="cfg-health-batch" min="10" max="100" step="10">
        </div>
      </div>
    </div>

    <!-- 地理过滤 -->
    <div class="form-section">
      <div class="form-section-title" data-i18n="config.section_geo_filter">地理过滤</div>
      <div class="form-grid">
        <div class="form-group">
          <label data-i18n="config.allowed_countries">允许国家（白名单）</label>
          <input type="text" id="cfg-allowed-countries" placeholder="US,JP,KR,SG">
          <div class="form-help" data-i18n="config.allowed_countries_help">非空时仅允许这些国家，忽略黑名单</div>
        </div>
        <div class="form-group">
          <label data-i18n="config.blocked_countries">屏蔽国家（黑名单）</label>
          <input type="text" id="cfg-blocked-countries" placeholder="CN,RU,KP">
          <div class="form-help" data-i18n="config.blocked_countries_help">白名单为空时生效</div>
        </div>
      </div>
    </div>

    <div class="modal-actions">
      <button class="btn btn-secondary" onclick="closeSettings()" data-i18n="config.cancel">取消</button>
      <button class="btn" onclick="saveConfig()" data-i18n="config.save">保存配置</button>
    </div>
  </div>
</div>

<!-- 添加订阅弹窗 -->
<div class="modal-overlay" id="sub-modal" onclick="if(event.target===this) closeSubModal()" style="display:none">
  <div class="modal" style="max-width:500px">
    <div class="modal-title" data-i18n="sub.add_title">添加订阅</div>
    <div class="form-section">
      <div class="form-grid">
        <div class="form-group">
          <label data-i18n="sub.name">名称</label>
          <input type="text" id="sub-name" placeholder="">
        </div>
        <div class="form-group" style="grid-column:1/-1">
          <label data-i18n="sub.import_mode">导入方式</label>
          <div style="display:flex;gap:8px;margin-bottom:8px">
            <button id="tab-url" class="ctrl-btn-primary" onclick="switchSubTab('url')" style="flex:1" data-i18n="sub.tab_url">订阅 URL</button>
            <button id="tab-file" class="ctrl-btn-secondary" onclick="switchSubTab('file')" style="flex:1" data-i18n="sub.tab_file">上传文件</button>
          </div>
        </div>
        <div class="form-group" id="sub-url-group" style="grid-column:1/-1">
          <label data-i18n="sub.url_label">订阅 URL</label>
          <input type="text" id="sub-url" placeholder="https://example.com/sub?token=xxx">
          <div class="form-help" data-i18n="sub.url_help">自动识别格式</div>
        </div>
        <div class="form-group" id="sub-file-group" style="grid-column:1/-1;display:none">
          <label data-i18n="sub.file_label">配置文件</label>
          <div style="border:1px dashed var(--border);padding:16px;text-align:center;cursor:pointer;transition:all 0.2s"
               onclick="document.getElementById('sub-file-input').click()"
               ondragover="event.preventDefault();this.style.borderColor='var(--fg)'"
               ondragleave="this.style.borderColor='var(--border)'"
               ondrop="event.preventDefault();this.style.borderColor='var(--border)';handleFileDrop(event)">
            <div id="sub-file-label" style="color:var(--fg-dim);font-size:11px" data-i18n="sub.file_drop">点击选择或拖拽文件到此处</div>
          </div>
          <input type="file" id="sub-file-input" accept=".yaml,.yml,.txt,.conf,.json" style="display:none" onchange="handleFileSelect(this)">
        </div>
        <div class="form-group">
          <label data-i18n="sub.refresh_min">刷新间隔 (分钟)</label>
          <input type="number" id="sub-refresh" value="60" min="10" max="1440" step="10">
          <div class="form-help" data-i18n="sub.refresh_min_help">仅 URL 模式有效</div>
        </div>
      </div>
    </div>
    <div class="modal-actions">
      <button class="btn btn-secondary" onclick="closeSubModal()" data-i18n="sub.cancel">取消</button>
      <button class="btn" onclick="addSubscription()" data-i18n="sub.submit">添加</button>
    </div>
  </div>
</div>

<!-- 手工新增节点弹窗 -->
<div class="modal-overlay" id="manual-modal" onclick="if(event.target===this) closeManualModal()" style="display:none">
  <div class="modal" style="max-width:620px">
    <div class="modal-title" data-i18n="proxy.add_manual_title">新增节点</div>
    <div class="form-section">
      <div class="form-group" style="grid-column:1/-1">
        <label data-i18n="proxy.manual_input">节点列表</label>
        <textarea id="manual-content" style="min-height:260px;padding:12px;background:var(--bg-card);border:1px solid var(--border);font-size:12px;font-family:var(--mono);color:var(--fg);outline:none;resize:vertical;line-height:1.6" placeholder="socks5://user:pass@192.168.1.1:1080&#10;http://192.168.1.1:8080&#10;https://user:pass@proxy.example.com:443"></textarea>
        <div class="form-help" data-i18n="proxy.manual_help">每行输入一个代理，支持 socks5://user:pass@host:port、http://host:port、https://user:pass@host:port</div>
      </div>
    </div>
    <div class="modal-actions">
      <button class="btn btn-secondary" onclick="closeManualModal()" data-i18n="sub.cancel">取消</button>
      <button class="btn" id="manual-submit-btn" onclick="submitManualNodes()" data-i18n="proxy.add_manual_submit">添加</button>
    </div>
  </div>
</div>

<!-- 访客贡献订阅弹窗 -->
<div class="modal-overlay" id="contribute-modal" onclick="if(event.target===this) closeContributeModal()" style="display:none">
  <div class="modal" style="max-width:460px">
    <div class="modal-title" data-i18n="contribute.title">贡献订阅</div>
    <div style="color:var(--fg-dim);font-size:11px;margin-bottom:16px;line-height:1.6">
      <span data-i18n="contribute.desc">分享你的代理订阅，帮助丰富代理池。</span><br>
      <span style="color:var(--gray-5);font-size:10px" data-i18n="contribute.privacy">你的订阅仅用于此代理池，不会被用于其他渠道。连续探测无可用节点将自动移除。</span>
    </div>
    <div class="form-section">
      <div class="form-grid">
        <div class="form-group" style="grid-column:1/-1">
          <label data-i18n="sub.name">名称</label>
          <input type="text" id="contribute-name" placeholder="">
        </div>
        <div class="form-group" style="grid-column:1/-1">
          <label data-i18n="sub.import_mode">导入方式</label>
          <div style="display:flex;gap:8px;margin-bottom:8px">
            <button id="ctab-url" class="ctrl-btn-primary" onclick="switchContributeTab('url')" style="flex:1" data-i18n="sub.tab_url">订阅 URL</button>
            <button id="ctab-file" class="ctrl-btn-secondary" onclick="switchContributeTab('file')" style="flex:1" data-i18n="sub.tab_file">上传文件</button>
          </div>
        </div>
        <div class="form-group" id="contribute-url-group" style="grid-column:1/-1">
          <label data-i18n="sub.url_label">订阅 URL</label>
          <input type="text" id="contribute-url" placeholder="https://example.com/sub?token=xxx">
          <div class="form-help" data-i18n="sub.url_help">自动识别格式</div>
        </div>
        <div class="form-group" id="contribute-file-group" style="grid-column:1/-1;display:none">
          <label data-i18n="sub.file_label">配置文件</label>
          <div style="border:1px dashed var(--border);padding:16px;text-align:center;cursor:pointer;transition:all 0.2s"
               onclick="document.getElementById('contribute-file-input').click()"
               ondragover="event.preventDefault();this.style.borderColor='var(--fg)'"
               ondragleave="this.style.borderColor='var(--border)'"
               ondrop="event.preventDefault();this.style.borderColor='var(--border)';handleContributeFileDrop(event)">
            <div id="contribute-file-label" style="color:var(--fg-dim);font-size:11px" data-i18n="sub.file_drop">点击选择或拖拽文件到此处</div>
          </div>
          <input type="file" id="contribute-file-input" accept=".yaml,.yml,.txt,.conf,.json" style="display:none" onchange="handleContributeFileSelect(this)">
        </div>
      </div>
    </div>
    <div class="modal-actions">
      <button class="btn btn-secondary" onclick="closeContributeModal()" data-i18n="sub.cancel">取消</button>
      <button class="btn" id="contribute-submit-btn" onclick="submitContribution()" data-i18n="contribute.submit">提交</button>
    </div>
  </div>
</div>

<script>
// 国际化翻译
const i18n = {
  zh: {
    'nav.config': '配置',
    'nav.login': '登录',
    'nav.logout': '退出',
    'health.status': '池子状态',
    'health.total': '总代理数',
    'health.capacity': '容量',
    'health.slots': '槽位',
    'health.avg': '平均',
    'health.state.healthy': '健康',
    'health.state.warning': '警告',
    'health.state.critical': '危急',
    'health.state.emergency': '紧急',
    'quality.title': '质量分布',
    'quality.grade_s': 'S级',
    'quality.grade_a': 'A级',
    'quality.grade_b': 'B级',
    'quality.grade_c': 'C级',
    'actions.fetch': '抓取代理',
    'actions.refresh': '刷新延迟',
    'actions.config': '配置池子',
    'control.proxy_config': '代理配置地址',
    'control.http_random': 'HTTP随机',
    'control.http_stable': 'HTTP稳定',
    'control.socks_random': 'SOCKS随机',
    'control.socks_stable': 'SOCKS稳定',
    'proxy.title': '代理列表',
    'proxy.tab_all': '全部',
    'proxy.filter_protocol': '协议',
    'proxy.filter_country': '出口国家',
    'proxy.loading': '加载中...',
    'proxy.empty': '暂无代理',
    'proxy.th_grade': '等级',
    'proxy.th_protocol': '协议',
    'proxy.th_address': '地址',
    'proxy.th_exit_ip': '出口IP',
    'proxy.th_location': '位置',
    'proxy.th_ip_type': 'IP类型',
    'proxy.th_risk': '风险',
    'proxy.th_residential': '住宅',
    'proxy.th_latency': '延迟',
    'proxy.th_usage': '使用统计',
    'proxy.th_action': '操作',
    'proxy.yes': '是',
    'proxy.no': '否',
    'proxy.btn_delete': '删除',
    'proxy.btn_refresh': '刷新',
    'proxy.add_manual': '新增节点',
    'proxy.add_manual_title': '新增节点',
    'proxy.add_manual_submit': '添加',
    'proxy.add_manual_submitting': '添加中...',
    'proxy.manual_input': '节点列表',
    'proxy.manual_help': '每行输入一个代理，支持 socks5://user:pass@host:port、http://host:port、https://user:pass@host:port',
    'proxy.copy_success': '已复制',
    'proxy.refresh_started': '刷新已启动',
    'log.title': '系统日志',
    'log.auto_refresh_label': '自动刷新',
    'log.loading': '加载中...',
    'log.empty': '暂无日志',
    'config.title': '池子配置',
    'config.section_capacity': '池子容量',
    'config.max_size': '最大容量',
    'config.max_size_help': '代理池总槽位数',
    'config.http_ratio': 'HTTP占比',
    'config.http_ratio_help': '0.5 = 50% HTTP, 50% SOCKS5',
    'config.min_per_protocol': '每协议最小数',
    'config.min_per_protocol_help': '最小保证数量',
    'config.section_latency': '延迟标准 (ms)',
    'config.latency_standard': '标准模式',
    'config.latency_healthy': '健康模式',
    'config.latency_emergency': '紧急模式',
    'config.section_validation': '验证与健康检查',
    'config.validate_concurrency': '验证并发数',
    'config.validate_timeout': '验证超时(秒)',
    'config.health_interval': '检查间隔(分钟)',
    'config.health_batch': '每批数量',
    'config.section_optimization': '优化设置',
    'config.optimize_interval': '优化间隔(分钟)',
    'config.replace_threshold': '替换阈值',
    'config.replace_threshold_help': '新代理需快30%',
    'config.section_geo_filter': '地理过滤',
    'config.allowed_countries': '允许国家（白名单）',
    'config.allowed_countries_help': '非空时仅允许这些国家入池，忽略黑名单',
    'config.blocked_countries': '屏蔽国家（黑名单）',
    'config.blocked_countries_help': '白名单为空时生效',
    'config.cancel': '取消',
    'config.save': '保存配置',
    'msg.fetch_confirm': '确定开始抓取代理吗？',
    'msg.fetch_started': '抓取已在后台启动',
    'msg.refresh_confirm': '确定刷新所有代理的延迟吗？这可能需要一些时间。',
    'msg.refresh_started': '延迟刷新已启动',
    'msg.delete_confirm': '确定删除代理',
    'msg.config_saved': '配置保存成功',
    'msg.config_failed': '配置保存失败',
    // 设置弹窗新增
    'config.system_title': '系统设置',
    'config.section_proxy_mode': '代理使用模式',
    'config.proxy_strategy': '出站代理选择策略',
    'config.mode_mixed_custom': '混合 · 订阅优先（有订阅代理时优先使用，无可用则降级到免费）',
    'config.mode_mixed_free': '混合 · 免费优先（有免费代理时优先使用，无可用则降级到订阅）',
    'config.mode_mixed': '混合 · 平等（不区分来源，按延迟/随机选择）',
    'config.mode_custom_only': '仅订阅代理（只使用订阅导入的代理）',
    'config.mode_free_only': '仅免费代理（只使用公开抓取的代理）',
    'config.section_free_pool': '免费代理池',
    'config.pool_capacity': '池子容量',
    'config.pool_capacity_help': '免费代理总槽位',
    'config.http_ratio_label': 'HTTP 占比',
    'config.latency_standard': '标准延迟 (ms)',
    'config.latency_healthy': '健康延迟 (ms)',
    'config.latency_emergency': '紧急延迟 (ms)',
    'config.section_sub_pool': '订阅代理池',
    'config.probe_interval': '探测间隔 (分钟)',
    'config.probe_interval_help': '禁用代理的唤醒探测间隔',
    'config.refresh_interval': '默认刷新间隔 (分钟)',
    'config.refresh_interval_help': '新订阅的默���刷新周期',
    'config.geo_filter_help': '免费代理删除，订阅代理禁用',
    // 健康面板
    'health.free_pool': 'FREE_POOL',
    'health.sub_pool': 'SUBSCRIPTION_POOL',
    'health.free_proxies': '免费代理',
    'health.sub_sources': '订阅源',
    'health.available': '可用',
    'health.disabled': '禁用/待恢复',
    'health.awaiting_probe': '等待探测唤醒',
    'health.no_disabled': '无禁用节点',
    'health.singbox_running': 'sing-box 运行中',
    'health.ready': '就绪',
    'health.not_added': '未添加',
    'health.total_nodes': '共 {0} 节点',
    // 订阅面板
    'sub.title': 'SUBSCRIPTIONS',
    'sub.add': '添加订阅',
    'sub.refresh_all': '刷新所有订阅',
    'sub.empty': '暂无订阅',
    'sub.nodes': '节点',
    'sub.available': '可用',
    'sub.disabled_label': '禁用',
    'sub.contributed': '贡献',
    // 添加订阅弹窗
    'sub.add_title': '添加订阅',
    'sub.name': '名称',
    'sub.import_mode': '导入方式',
    'sub.tab_url': '订阅 URL',
    'sub.tab_file': '上传文件',
    'sub.url_label': '订阅 URL',
    'sub.url_help': '自动识别格式：Clash YAML / V2ray 链接 / Base64 / 纯文本',
    'sub.file_label': '配置文件',
    'sub.file_drop': '点击选择或拖拽文件到此处',
    'sub.file_formats': '支持 Clash YAML / V2ray 订阅 / 纯文本',
    'sub.refresh_min': '刷新间隔 (分钟)',
    'sub.refresh_min_help': '仅 URL 模式有效，上传文件不自动刷新',
    'sub.cancel': '取消',
    'sub.submit': '添加',
    // 贡献订阅弹窗
    'contribute.title': '贡献订阅',
    'contribute.desc': '分享你的代理订阅，帮助丰富代理池。',
    'contribute.privacy': '你的订阅仅用于此代理池，不会被用于其他渠道。连续探测无可用节点将自动移除。',
    'contribute.submit': '提交',
    'contribute.validating': '验证中...',
    'contribute.nav': '贡献订阅',
    'contribute.settings': '系统设置',
    // 消息
    'msg.sub_added': '订阅已添加，正在导入节点...',
    'msg.sub_refreshed': '刷新已启动',
    'msg.sub_refresh_all': '所有订阅刷新已启动',
    'msg.sub_delete_confirm': '确定删除此订阅？',
    'msg.sub_url_required': '请填写订阅 URL',
    'msg.sub_file_required': '请选择或拖拽配置文件',
    'msg.contribute_thanks': '感谢贡献！订阅已添加，正在导入节点...',
    'msg.manual_required': '请填写节点列表',
    'msg.manual_added': '已添加 {0} 个节点，跳过 {1} 个无效节点',
    'msg.submit_failed': '提交失败: ',
  },
  en: {
    'nav.config': 'Config',
    'nav.login': 'Login',
    'nav.logout': 'Logout',
    'health.status': 'Pool Status',
    'health.total': 'Total Proxies',
    'health.capacity': 'capacity',
    'health.slots': 'slots',
    'health.avg': 'avg',
    'health.state.healthy': 'Healthy',
    'health.state.warning': 'Warning',
    'health.state.critical': 'Critical',
    'health.state.emergency': 'Emergency',
    'quality.title': 'Quality Distribution',
    'quality.grade_s': 'S Grade',
    'quality.grade_a': 'A Grade',
    'quality.grade_b': 'B Grade',
    'quality.grade_c': 'C Grade',
    'actions.fetch': 'Fetch Proxies',
    'actions.refresh': 'Refresh Latency',
    'actions.config': 'Configure Pool',
    'control.proxy_config': 'Proxy Config Addresses',
    'control.http_random': 'HTTP Random',
    'control.http_stable': 'HTTP Stable',
    'control.socks_random': 'SOCKS Random',
    'control.socks_stable': 'SOCKS Stable',
    'proxy.title': 'Proxy Registry',
    'proxy.tab_all': 'All',
    'proxy.filter_protocol': 'Protocol',
    'proxy.filter_country': 'Exit Country',
    'proxy.loading': 'Loading...',
    'proxy.empty': 'No proxies available',
    'proxy.th_grade': 'Grade',
    'proxy.th_protocol': 'Protocol',
    'proxy.th_address': 'Address',
    'proxy.th_exit_ip': 'Exit IP',
    'proxy.th_location': 'Location',
    'proxy.th_ip_type': 'IP Type',
    'proxy.th_risk': 'Risk',
    'proxy.th_residential': 'Residential',
    'proxy.th_latency': 'Latency',
    'proxy.th_usage': 'Usage',
    'proxy.th_action': 'Action',
    'proxy.yes': 'Yes',
    'proxy.no': 'No',
    'proxy.btn_delete': 'DEL',
    'proxy.btn_refresh': 'Refresh',
    'proxy.add_manual': 'Add Node',
    'proxy.add_manual_title': 'Add Node',
    'proxy.add_manual_submit': 'Add',
    'proxy.add_manual_submitting': 'Adding...',
    'proxy.manual_input': 'Node List',
    'proxy.manual_help': 'One proxy per line. Supports socks5://user:pass@host:port, http://host:port, https://user:pass@host:port',
    'proxy.copy_success': 'Copied',
    'proxy.refresh_started': 'Refresh started',
    'log.title': 'System Log',
    'log.auto_refresh_label': 'Auto Refresh',
    'log.loading': 'Loading...',
    'log.empty': 'No logs',
    'config.title': 'Pool Configuration',
    'config.section_capacity': 'Pool Capacity',
    'config.max_size': 'Max Size',
    'config.max_size_help': 'Total proxy slots',
    'config.http_ratio': 'HTTP Ratio',
    'config.http_ratio_help': '0.5 = 50% HTTP, 50% SOCKS5',
    'config.min_per_protocol': 'Min Per Protocol',
    'config.min_per_protocol_help': 'Minimum guarantee',
    'config.section_latency': 'Latency Standards (ms)',
    'config.latency_standard': 'Standard',
    'config.latency_healthy': 'Healthy',
    'config.latency_emergency': 'Emergency',
    'config.section_validation': 'Validation & Health Check',
    'config.validate_concurrency': 'Validate Concurrency',
    'config.validate_timeout': 'Validate Timeout (s)',
    'config.health_interval': 'Health Check Interval (min)',
    'config.health_batch': 'Batch Size',
    'config.section_optimization': 'Optimization',
    'config.optimize_interval': 'Optimize Interval (min)',
    'config.replace_threshold': 'Replace Threshold',
    'config.replace_threshold_help': 'New proxy must be 30% faster',
    'config.section_geo_filter': 'Geo Filter',
    'config.allowed_countries': 'Allowed Countries (Whitelist)',
    'config.allowed_countries_help': 'When set, only these countries are allowed; blacklist is ignored',
    'config.blocked_countries': 'Blocked Countries (Blacklist)',
    'config.blocked_countries_help': 'Effective only when whitelist is empty',
    'config.cancel': 'Cancel',
    'config.save': 'Save Configuration',
    'msg.fetch_confirm': 'Start proxy fetch?',
    'msg.fetch_started': 'Fetch started in background',
    'msg.refresh_confirm': 'Refresh latency for all proxies? This may take a while.',
    'msg.refresh_started': 'Latency refresh started',
    'msg.delete_confirm': 'Delete proxy',
    'msg.config_saved': 'Configuration saved successfully',
    'msg.config_failed': 'Failed to save configuration',
    'config.system_title': 'System Settings',
    'config.section_proxy_mode': 'Proxy Mode',
    'config.proxy_strategy': 'Outbound Proxy Strategy',
    'config.mode_mixed_custom': 'Mixed · Subscription Priority',
    'config.mode_mixed_free': 'Mixed · Free Priority',
    'config.mode_mixed': 'Mixed · Equal (select by latency/random)',
    'config.mode_custom_only': 'Subscription Only',
    'config.mode_free_only': 'Free Only',
    'config.section_free_pool': 'Free Proxy Pool',
    'config.pool_capacity': 'Pool Capacity',
    'config.pool_capacity_help': 'Total free proxy slots',
    'config.http_ratio_label': 'HTTP Ratio',
    'config.latency_standard': 'Standard Latency (ms)',
    'config.latency_healthy': 'Healthy Latency (ms)',
    'config.latency_emergency': 'Emergency Latency (ms)',
    'config.section_sub_pool': 'Subscription Pool',
    'config.probe_interval': 'Probe Interval (min)',
    'config.probe_interval_help': 'Wake-up probe interval for disabled proxies',
    'config.refresh_interval': 'Default Refresh (min)',
    'config.refresh_interval_help': 'Default refresh cycle for new subscriptions',
    'config.geo_filter_help': 'Free: delete, Subscription: disable',
    'health.free_pool': 'FREE_POOL',
    'health.sub_pool': 'SUBSCRIPTION_POOL',
    'health.free_proxies': 'Free Proxies',
    'health.sub_sources': 'Sources',
    'health.available': 'Available',
    'health.disabled': 'Disabled',
    'health.awaiting_probe': 'Awaiting probe',
    'health.no_disabled': 'No disabled nodes',
    'health.singbox_running': 'sing-box running',
    'health.ready': 'Ready',
    'health.not_added': 'None',
    'health.total_nodes': '{0} total nodes',
    'sub.title': 'SUBSCRIPTIONS',
    'sub.add': 'Add Subscription',
    'sub.refresh_all': 'Refresh All',
    'sub.empty': 'No subscriptions',
    'sub.nodes': 'nodes',
    'sub.available': 'available',
    'sub.disabled_label': 'disabled',
    'sub.contributed': 'Contributed',
    'sub.add_title': 'Add Subscription',
    'sub.name': 'Name',
    'sub.import_mode': 'Import Mode',
    'sub.tab_url': 'URL',
    'sub.tab_file': 'Upload File',
    'sub.url_label': 'Subscription URL',
    'sub.url_help': 'Auto-detect: Clash YAML / V2ray / Base64 / Plain text',
    'sub.file_label': 'Config File',
    'sub.file_drop': 'Click or drag file here',
    'sub.file_formats': 'Supports Clash YAML / V2ray / Plain text',
    'sub.refresh_min': 'Refresh Interval (min)',
    'sub.refresh_min_help': 'URL mode only; file uploads do not auto-refresh',
    'sub.cancel': 'Cancel',
    'sub.submit': 'Add',
    'contribute.title': 'Contribute Subscription',
    'contribute.desc': 'Share your proxy subscription to enrich the pool.',
    'contribute.privacy': 'Your subscription is only used for this proxy pool. Subscriptions with no available nodes for 7 days will be auto-removed.',
    'contribute.submit': 'Submit',
    'contribute.validating': 'Validating...',
    'contribute.nav': 'Contribute',
    'contribute.settings': 'Settings',
    'msg.sub_added': 'Subscription added, importing nodes...',
    'msg.sub_refreshed': 'Refresh started',
    'msg.sub_refresh_all': 'Refreshing all subscriptions',
    'msg.sub_delete_confirm': 'Delete this subscription?',
    'msg.sub_url_required': 'Please enter subscription URL',
    'msg.sub_file_required': 'Please select or drag a config file',
    'msg.contribute_thanks': 'Thanks! Subscription added, importing nodes...',
    'msg.manual_required': 'Please enter proxy lines',
    'msg.manual_added': 'Added {0} nodes, skipped {1} invalid nodes',
    'msg.submit_failed': 'Submit failed: ',
  }
};

let currentLang = 'zh';
let logCountdown = 5;

function t(key) {
  return i18n[currentLang][key] || key;
}

function updateLogCountdown() {
  const el = document.getElementById('log-countdown');
  if (el) el.textContent = logCountdown;
}

function updateI18n() {
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    el.textContent = t(key);
  });
  // 更新 title 属性
  document.querySelectorAll('[data-i18n-title]').forEach(el => {
    const key = el.getAttribute('data-i18n-title');
    el.title = t(key);
  });
  document.getElementById('lang-btn').textContent = currentLang === 'zh' ? 'EN' : '中';
  document.title = currentLang === 'zh' ? 'GoProxy — 智能代理池' : 'GoProxy — Intelligent Pool';

  // 更新筛选下拉框标签
  const protocolLabel = document.getElementById('protocol-filter-label');
  if (protocolLabel) protocolLabel.textContent = t('proxy.filter_protocol');
  const countryLabel = document.getElementById('country-filter-label');
  if (countryLabel) countryLabel.textContent = t('proxy.filter_country');
}

function toggleLang() {
  currentLang = currentLang === 'zh' ? 'en' : 'zh';
  document.getElementById('lang-btn').textContent = currentLang === 'zh' ? '[ EN ]' : '[ 中文 ]';
  localStorage.setItem('lang', currentLang);
  updateI18n();
  if (allProxies.length > 0) {
    filterAndRender();
  }
  // 重新渲染包含动态 t() 文字的模块
  loadSubscriptions();
  loadPoolStatus();
  loadProxyConfig();
}

// 页面加载时恢复语言设置
const savedLang = localStorage.getItem('lang');
if (savedLang) {
  currentLang = savedLang;
  updateI18n();
}

let currentProtocol = '';
let currentCountry = '';
let allProxies = [];
let isAdmin = false; // 是否为管理员

async function api(path, opts) {
  const r = await fetch(path, opts);
  if (r.status === 401) { location.href = '/login'; return null; }
  return r.json();
}

// 检查当前用户权限
async function checkAuth() {
  try {
    const auth = await fetch('/api/auth/check').then(r => r.json());
    isAdmin = auth.isAdmin || false;
    updateUIByRole();
  } catch (e) {
    isAdmin = false;
    updateUIByRole();
  }
}

// 根据角色更新 UI
function updateUIByRole() {
  // 显示/隐藏管理员专属元素
  document.querySelectorAll('.admin-only').forEach(el => {
    if (isAdmin) {
      el.style.display = 'block';
    } else {
      el.style.display = 'none';
    }
  });
  
  // 显示/隐藏登录链接和访客专属元素
  const loginLink = document.getElementById('login-link');
  if (loginLink) loginLink.style.display = isAdmin ? 'none' : 'inline-flex';
  document.querySelectorAll('.guest-only').forEach(el => {
    el.style.display = isAdmin ? 'none' : 'inline-flex';
  });
  
  // 更新用户模式标识
  const modeEl = document.getElementById('user-mode');
  if (modeEl) {
    if (isAdmin) {
      modeEl.textContent = 'admin';
    } else {
      modeEl.textContent = 'guest';
    }
  }
  
  // 重新渲染代理列表（更新操作列）
  if (allProxies.length > 0) {
    filterAndRender();
  }
}

function getCountryFlag(countryCode) {
  if (!countryCode || countryCode === 'UNKNOWN') return '';
  const offset = 127397;
  return countryCode.toUpperCase().split('').map(c => String.fromCodePoint(c.charCodeAt(0) + offset)).join('');
}

function esc(value) {
  return String(value ?? '').replace(/[&<>"']/g, ch => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  }[ch]));
}

function jsString(value) {
  return String(value ?? '').replace(/\\/g, '\\\\').replace(/'/g, "\\'");
}

function showToast(message) {
  const toast = document.getElementById('toast');
  toast.textContent = message;
  toast.classList.add('show');
  setTimeout(() => toast.classList.remove('show'), 2000);
}

function copyToClipboard(text) {
  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard.writeText(text).then(() => {
      showToast(t('proxy.copy_success') + ': ' + text);
    }).catch(err => {
      console.error('Copy failed:', err);
    });
    return;
  }
  const input = document.createElement('textarea');
  input.value = text;
  input.style.position = 'fixed';
  input.style.opacity = '0';
  document.body.appendChild(input);
  input.focus();
  input.select();
  try {
    document.execCommand('copy');
    showToast(t('proxy.copy_success') + ': ' + text);
  } catch (err) {
    console.error('Copy failed:', err);
  }
  document.body.removeChild(input);
}

function normalizePort(portValue) {
  return String(portValue || '').replace(':', '');
}

function proxyHost() {
  return window.location.hostname || window.location.host.split(':')[0];
}

function renderProxyConfig(cfg) {
  const el = document.getElementById('proxy-config-list');
  if (!el || !cfg) return;

  const host = proxyHost();
  const items = [
    {label: t('control.http_random'), value: 'http://' + host + ':' + normalizePort(cfg.proxy_port)},
    {label: t('control.http_stable'), value: 'http://' + host + ':' + normalizePort(cfg.stable_proxy_port)},
    {label: t('control.socks_random'), value: 'socks5://' + host + ':' + normalizePort(cfg.socks5_port)},
    {label: t('control.socks_stable'), value: 'socks5://' + host + ':' + normalizePort(cfg.stable_socks5_port)}
  ];

  el.innerHTML = items.map(item =>
    '<div class="proxy-config-row">' +
      '<span class="proxy-config-label">' + item.label + '</span>' +
      '<span class="proxy-config-address" title="' + item.value + '">' + item.value + '</span>' +
      '<button class="proxy-copy-btn" type="button" onclick="copyToClipboard(\'' + item.value + '\')" title="Copy">⧉</button>' +
    '</div>'
  ).join('');
}

async function loadProxyConfig() {
  const cfg = await api('/api/config');
  renderProxyConfig(cfg);
}

async function refreshProxy(address) {
  const res = await api('/api/proxy/refresh', { address });
  if (res) {
    showToast(t('proxy.refresh_started'));
    setTimeout(() => loadProxies(currentFilter), 2000);
  }
}

async function loadPoolStatus() {
  const status = await api('/api/pool/status');
  if (!status) return;

  const freeTotal = status.Total - (status.CustomCount || 0);
  document.getElementById('stat-total').textContent = freeTotal;
  document.getElementById('stat-capacity').textContent = status.HTTPSlots + status.SOCKS5Slots;
  document.getElementById('stat-http').textContent = status.HTTP;
  document.getElementById('stat-socks5').textContent = status.SOCKS5;
  document.getElementById('http-slots').textContent = status.HTTPSlots;
  document.getElementById('socks5-slots').textContent = status.SOCKS5Slots;
  document.getElementById('http-avg').textContent = status.AvgLatencyHTTP || '—';
  document.getElementById('socks5-avg').textContent = status.AvgLatencySocks5 || '—';
  
  const stateEl = document.getElementById('pool-state');
  const dotEl = document.getElementById('pool-status-dot');
  const stateText = t('health.state.' + status.State.toLowerCase());
  stateEl.textContent = stateText.toUpperCase();
  dotEl.className = 'health-status ' + status.State.toLowerCase();
}

async function loadQualityDistribution() {
  const dist = await api('/api/pool/quality');
  if (!dist) return;

  const total = (dist.S || 0) + (dist.A || 0) + (dist.B || 0) + (dist.C || 0);
  
  document.getElementById('grade-s-count').textContent = dist.S || 0;
  document.getElementById('grade-a-count').textContent = dist.A || 0;
  document.getElementById('grade-b-count').textContent = dist.B || 0;
  document.getElementById('grade-c-count').textContent = dist.C || 0;

  if (total > 0) {
    const visual = document.getElementById('quality-visual');
    visual.innerHTML = '';
    if (dist.S) visual.innerHTML += '<div class="quality-segment quality-s" style="width:' + (dist.S/total*100) + '%">' + (dist.S/total*100 >= 10 ? 'S' : '') + '</div>';
    if (dist.A) visual.innerHTML += '<div class="quality-segment quality-a" style="width:' + (dist.A/total*100) + '%">' + (dist.A/total*100 >= 10 ? 'A' : '') + '</div>';
    if (dist.B) visual.innerHTML += '<div class="quality-segment quality-b" style="width:' + (dist.B/total*100) + '%">' + (dist.B/total*100 >= 10 ? 'B' : '') + '</div>';
    if (dist.C) visual.innerHTML += '<div class="quality-segment quality-c" style="width:' + (dist.C/total*100) + '%">' + (dist.C/total*100 >= 10 ? 'C' : '') + '</div>';
  }
}

let subNameMap = {};
async function loadProxies() {
  // 先加载订阅名称映射
  const subs = await api('/api/subscriptions');
  if (subs) {
    subNameMap = {};
    subs.forEach(s => { subNameMap[s.id] = s.name || t('sub.add_title'); });
  }

  const path = currentProtocol ? '/api/proxies?protocol=' + currentProtocol : '/api/proxies';
  const proxies = await api(path);
  if (!proxies) return;

  allProxies = proxies;
  updateCountryOptions();
  filterAndRender();
}

function updateCountryOptions() {
  const countries = new Set();
  allProxies.forEach(p => {
    if (p.exit_location) {
      const countryCode = p.exit_location.split(' ')[0];
      if (countryCode) countries.add(countryCode);
    }
  });
  
  const select = document.getElementById('country-filter');
  const currentValue = select.value;
  select.innerHTML = '<option value="" id="country-filter-label">' + t('proxy.filter_country') + '</option>';
  Array.from(countries).sort().forEach(code => {
    const flag = getCountryFlag(code);
    select.innerHTML += '<option value="' + code + '">' + flag + ' ' + code + '</option>';
  });
  if (currentValue && countries.has(currentValue)) {
    select.value = currentValue;
  }
}

function filterAndRender() {
  let filtered = allProxies;
  if (currentCountry) {
    filtered = filtered.filter(p => p.exit_location && (p.exit_location === currentCountry || p.exit_location.startsWith(currentCountry + ' ')));
  }
  renderProxies(filtered);
}

function setProtocolFilter(protocol) {
  currentProtocol = protocol;
  loadProxies();
}

function setCountryFilter(country) {
  currentCountry = country;
  filterAndRender();
}

function renderRiskBadge(p) {
  const level = p.risk_level || 'Unknown';
  if (!level || level === 'Unknown') {
    return '<span class="badge badge-muted">—</span>';
  }
  const score = Number(p.risk_score || 0);
  const cls = score < 0.25 ? 'badge-risk-low' : score < 0.5 ? 'badge-risk-med' : 'badge-risk-high';
  return '<span class="badge ' + cls + '">' + esc(level) + ' ' + Math.round(score * 100) + '%</span>';
}

function hasIPProfile(p) {
  return !!(p.ip_type || (p.risk_level && p.risk_level !== 'Unknown'));
}

function renderIPTypeLine(p) {
  return '<div class="cell-sub">' + t('proxy.th_ip_type') + ': ' + (p.ip_type ? esc(p.ip_type) : '—') + '</div>';
}

function renderResidentialLine(p) {
  if (!hasIPProfile(p)) {
    return '<div class="cell-sub">' + t('proxy.th_residential') + ': —</div>';
  }
  return '<div class="cell-sub">' + t('proxy.th_residential') + ': ' + (p.is_residential ? t('proxy.yes') : t('proxy.no')) + '</div>';
}

function renderProxies(proxies) {
  let html = '';
  if (proxies.length === 0) {
    html = '<div class="empty" data-i18n="proxy.empty">' + t('proxy.empty') + '</div>';
  } else {
    html = '<table><thead><tr>';
    html += '<th class="th-grade" data-i18n="proxy.th_grade">' + t('proxy.th_grade') + '</th>';
    html += '<th class="th-protocol" data-i18n="proxy.th_protocol">' + t('proxy.th_protocol') + '</th>';
    html += '<th class="th-address" data-i18n="proxy.th_address">' + t('proxy.th_address') + '</th>';
    html += '<th class="th-exit-ip" data-i18n="proxy.th_exit_ip">' + t('proxy.th_exit_ip') + '</th>';
    html += '<th class="th-location" data-i18n="proxy.th_location">' + t('proxy.th_location') + '</th>';
    html += '<th class="th-risk" data-i18n="proxy.th_risk">' + t('proxy.th_risk') + '</th>';
    html += '<th class="th-latency" data-i18n="proxy.th_latency">' + t('proxy.th_latency') + '</th>';
    html += '<th class="th-usage" data-i18n="proxy.th_usage">' + t('proxy.th_usage') + '</th>';
    if (isAdmin) {
      html += '<th class="th-action" data-i18n="proxy.th_action">' + t('proxy.th_action') + '</th>';
    }
    html += '</tr></thead><tbody>';

    proxies.forEach(p => {
      const flag = p.exit_location ? getCountryFlag(p.exit_location.split(' ')[0]) : '';
      const grade = (p.quality_grade || 'C').toLowerCase();
      const latencyClass = 'grade-' + grade;
      
      const rowStyle = p.source === 'custom' ? ' style="border-left:2px solid var(--yellow)"' : '';
      html += '<tr' + rowStyle + '>';
      html += '<td class="cell-grade grade-' + grade + '">' + (p.quality_grade || 'C') + '</td>';
      html += '<td><div class="cell-protocol"><span class="badge badge-' + p.protocol + '">' + p.protocol.toUpperCase() + '</span>';
      if (p.source === 'custom') {
        const subName = subNameMap[p.subscription_id] || t('sub.add_title');
        html += '<span class="subscription-badge" title="' + esc(subName) + '">' + esc(subName) + '</span>';
      }
      html += '</div></td>';
      html += '<td class="cell-mono cell-clickable" onclick="copyToClipboard(\'' + jsString(p.address) + '\')" title="Copy">' + esc(p.address) + '</td>';
      html += '<td><div class="cell-stack"><div class="cell-main cell-mono">' + (p.exit_ip ? esc(p.exit_ip) : '—') + '</div>' + renderIPTypeLine(p) + '</div></td>';
      html += '<td><div class="cell-stack"><div class="cell-main">' + flag + ' ' + (p.exit_location ? esc(p.exit_location) : '—') + '</div>' + renderResidentialLine(p) + '</div></td>';
      html += '<td>' + renderRiskBadge(p) + '</td>';
      html += '<td class="cell-mono ' + latencyClass + '">' + (p.latency > 0 ? p.latency + 'ms' : '—') + '</td>';
      html += '<td class="cell-mono">' + (p.use_count || 0) + ' / ' + (p.success_count || 0) + '</td>';
      
      if (isAdmin) {
        html += '<td class="cell-action"><div class="action-stack">';
        html += '<button class="btn-action" onclick="refreshProxy(\'' + jsString(p.address) + '\')" data-i18n="proxy.btn_refresh">' + t('proxy.btn_refresh') + '</button>';
        html += '<button class="btn-danger" onclick="deleteProxy(\'' + jsString(p.address) + '\')" data-i18n="proxy.btn_delete">' + t('proxy.btn_delete') + '</button>';
        html += '</div></td>';
      }
      
      html += '</tr>';
    });

    html += '</tbody></table>';
  }

  document.getElementById('proxy-table-wrap').innerHTML = html;
}

async function triggerFetch() {
  if (!confirm(t('msg.fetch_confirm'))) return;
  await api('/api/fetch', {method: 'POST'});
  alert(t('msg.fetch_started'));
  setTimeout(loadAll, 2000);
}

async function refreshLatency() {
  if (!confirm(t('msg.refresh_confirm'))) return;
  await api('/api/refresh-latency', {method: 'POST'});
  alert(t('msg.refresh_started'));
  setTimeout(loadAll, 2000);
}

async function deleteProxy(addr) {
  if (!confirm(t('msg.delete_confirm') + ' ' + addr + '?')) return;
  await api('/api/proxy/delete', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({address: addr})
  });
  loadProxies();
}

async function loadLogs() {
  const data = await api('/api/logs');
  if (!data) return;
  
  const box = document.getElementById('logs-box');
  if (!data.lines || data.lines.length === 0) {
    box.innerHTML = '<div class="empty" data-i18n="log.empty">' + t('log.empty') + '</div>';
    return;
  }

  let html = '';
  data.lines.forEach(line => {
    let cls = '';
    if (line.includes('error') || line.includes('failed') || line.includes('❌') || line.includes('失败')) cls = 'error';
    if (line.includes('success') || line.includes('✅') || line.includes('completed') || line.includes('成功')) cls = 'success';
    html += '<div class="log-line ' + cls + '">' + line + '</div>';
  });
  box.innerHTML = html;
  box.scrollTop = box.scrollHeight;
  
  // 重置倒计时
  logCountdown = 5;
  
  // 同时刷新代理列表
  loadProxies();
}

async function openSettings() {
  const cfg = await api('/api/config');
  if (!cfg) return;

  document.getElementById('cfg-pool-size').value = cfg.pool_max_size;
  document.getElementById('cfg-http-ratio').value = cfg.pool_http_ratio;
  document.getElementById('cfg-min-per-protocol').value = cfg.pool_min_per_protocol;
  document.getElementById('cfg-max-latency').value = cfg.max_latency_ms;
  document.getElementById('cfg-max-latency-healthy').value = cfg.max_latency_healthy;
  document.getElementById('cfg-max-latency-emergency').value = cfg.max_latency_emergency;
  document.getElementById('cfg-concurrency').value = cfg.validate_concurrency;
  document.getElementById('cfg-timeout').value = cfg.validate_timeout;
  document.getElementById('cfg-health-interval').value = cfg.health_check_interval;
  document.getElementById('cfg-health-batch').value = cfg.health_check_batch_size;
  document.getElementById('cfg-optimize-interval').value = cfg.optimize_interval;
  document.getElementById('cfg-replace-threshold').value = cfg.replace_threshold;
  document.getElementById('cfg-blocked-countries').value = (cfg.blocked_countries || []).join(',');
  document.getElementById('cfg-allowed-countries').value = (cfg.allowed_countries || []).join(',');
  // 将 mode + priority 映射到5种模式
  const mode = cfg.custom_proxy_mode || 'mixed';
  const customPri = cfg.custom_priority === true;
  const freePri = cfg.custom_free_priority === true;
  let uiMode = 'mixed';
  if (mode === 'custom_only') uiMode = 'custom_only';
  else if (mode === 'free_only') uiMode = 'free_only';
  else if (mode === 'mixed' && customPri) uiMode = 'mixed_custom_priority';
  else if (mode === 'mixed' && freePri) uiMode = 'mixed_free_priority';
  else uiMode = 'mixed';
  document.getElementById('cfg-custom-mode').value = uiMode;
  document.getElementById('cfg-custom-probe').value = cfg.custom_probe_interval || 10;
  document.getElementById('cfg-custom-refresh').value = cfg.custom_refresh_interval || 60;

  document.getElementById('settings-modal').classList.add('show');
}

function closeSettings() {
  document.getElementById('settings-modal').classList.remove('show');
}

async function saveConfig() {
  const cfg = {
    pool_max_size: parseInt(document.getElementById('cfg-pool-size').value),
    pool_http_ratio: parseFloat(document.getElementById('cfg-http-ratio').value),
    pool_min_per_protocol: parseInt(document.getElementById('cfg-min-per-protocol').value),
    max_latency_ms: parseInt(document.getElementById('cfg-max-latency').value),
    max_latency_healthy: parseInt(document.getElementById('cfg-max-latency-healthy').value),
    max_latency_emergency: parseInt(document.getElementById('cfg-max-latency-emergency').value),
    validate_concurrency: parseInt(document.getElementById('cfg-concurrency').value),
    validate_timeout: parseInt(document.getElementById('cfg-timeout').value),
    health_check_interval: parseInt(document.getElementById('cfg-health-interval').value),
    health_check_batch_size: parseInt(document.getElementById('cfg-health-batch').value),
    optimize_interval: parseInt(document.getElementById('cfg-optimize-interval').value),
    replace_threshold: parseFloat(document.getElementById('cfg-replace-threshold').value),
    blocked_countries: document.getElementById('cfg-blocked-countries').value.split(',').map(s => s.trim().toUpperCase()).filter(s => s),
    allowed_countries: document.getElementById('cfg-allowed-countries').value.split(',').map(s => s.trim().toUpperCase()).filter(s => s),
    custom_proxy_mode: (() => {
      const m = document.getElementById('cfg-custom-mode').value;
      if (m === 'custom_only') return 'custom_only';
      if (m === 'free_only') return 'free_only';
      return 'mixed';
    })(),
    custom_priority: (() => {
      const m = document.getElementById('cfg-custom-mode').value;
      if (m === 'mixed_custom_priority') return true;
      if (m === 'mixed_free_priority') return false;
      return false;
    })(),
    custom_free_priority: document.getElementById('cfg-custom-mode').value === 'mixed_free_priority',
    custom_probe_interval: parseInt(document.getElementById('cfg-custom-probe').value),
    custom_refresh_interval: parseInt(document.getElementById('cfg-custom-refresh').value),
  };

  const result = await api('/api/config/save', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(cfg)
  });

  if (result && result.status === 'saved') {
    alert(t('msg.config_saved'));
    closeSettings();
    loadAll();
  } else {
    alert(t('msg.config_failed'));
  }
}

async function loadAll() {
  await checkAuth(); // 先检查权限
  loadPoolStatus();
  loadQualityDistribution();
  loadProxyConfig();
  loadProxies();
  loadLogs();
}

// ========== 订阅管理 ==========

async function loadSubscriptions() {
  const subs = await api('/api/subscriptions');
  const el = document.getElementById('sub-list');
  if (!el || !subs) return;

  if (subs.length === 0) {
    el.innerHTML = '<div style="color:var(--gray-5);text-align:center;padding:8px">' + t('sub.empty') + '</div>';
    return;
  }

  el.innerHTML = subs.map(s => {
    const statusColor = s.status === 'active' ? 'var(--green)' : 'var(--gray-5)';
    const statusIcon = s.status === 'active' ? '●' : '○';
    const active = s.active_count || 0;
    const disabled = s.disabled_count || 0;
    const total = active + disabled;
    const statsText = total + ' ' + t('sub.nodes') + ' · ' + active + ' ' + t('sub.available') + (disabled > 0 ? ' · ' + disabled + ' ' + t('sub.disabled_label') : '');
    const badge = s.contributed ? '<span style="display:inline-block;background:var(--orange);color:#000;font-size:7px;font-weight:700;padding:0 3px;margin-left:4px;vertical-align:middle">' + t('sub.contributed') + '</span>' : '';
    return '<div style="display:flex;align-items:center;justify-content:space-between;padding:4px 0;border-bottom:1px solid var(--border)">' +
      '<div style="flex:1;min-width:0">' +
        '<span style="color:' + statusColor + '">' + statusIcon + '</span> ' +
        '<span style="font-weight:600">' + (s.name||t('sub.add_title')) + '</span>' + badge +
        '<span style="color:var(--gray-5);margin-left:8px">' + statsText + '</span>' +
      '</div>' +
      '<div style="display:flex;gap:4px;flex-shrink:0">' +
        '<button onclick="refreshSub(' + s.id + ')" style="background:none;border:1px solid var(--border);color:var(--fg-dim);cursor:pointer;padding:2px 6px;font-size:9px;font-family:var(--mono)">↻</button>' +
        '<button onclick="toggleSub(' + s.id + ')" style="background:none;border:1px solid var(--border);color:var(--fg-dim);cursor:pointer;padding:2px 6px;font-size:9px;font-family:var(--mono)">' + (s.status === 'active' ? '⏸' : '▶') + '</button>' +
        '<button onclick="deleteSub(' + s.id + ')" style="background:none;border:1px solid var(--red);color:var(--red);cursor:pointer;padding:2px 6px;font-size:9px;font-family:var(--mono)">✕</button>' +
      '</div>' +
    '</div>';
  }).join('');

  // 加载状态
  const status = await api('/api/custom/status');
  const statusEl = document.getElementById('sub-status');
  if (status && statusEl) {
    const parts = [];
    if (status.singbox_running) parts.push('sing-box ✅ ' + status.singbox_nodes + ' ' + t('sub.nodes'));
    statusEl.textContent = parts.length > 0 ? parts.join(' · ') : '';
  }

  // 更新订阅代理统计卡片
  if (status) {
    const active = status.custom_count || 0;
    const disabled = status.disabled_count || 0;
    const subCount = status.subscription_count || 0;

    const subCountEl = document.getElementById('stat-sub-count');
    const subMetaEl = document.getElementById('stat-sub-meta');
    if (subCountEl) subCountEl.textContent = subCount;
    if (subMetaEl) subMetaEl.textContent = status.singbox_running ? t('health.singbox_running') : (subCount > 0 ? t('health.ready') : t('health.not_added'));

    const customEl = document.getElementById('stat-custom');
    const customMeta = document.getElementById('custom-meta');
    if (customEl) customEl.textContent = active;
    if (customMeta) customMeta.textContent = (active + disabled) > 0 ? t('health.total_nodes').replace('{0}', active + disabled) : '—';

    const disabledEl = document.getElementById('stat-custom-disabled');
    const disabledMeta = document.getElementById('custom-disabled-meta');
    if (disabledEl) disabledEl.textContent = disabled;
    if (disabledMeta) disabledMeta.textContent = disabled > 0 ? t('health.awaiting_probe') : t('health.no_disabled');
  }
}

let subFileContent = '';
let subTab = 'url';

function switchSubTab(tab) {
  subTab = tab;
  document.getElementById('sub-url-group').style.display = tab === 'url' ? '' : 'none';
  document.getElementById('sub-file-group').style.display = tab === 'file' ? '' : 'none';
  document.getElementById('tab-url').className = tab === 'url' ? 'ctrl-btn-primary' : 'ctrl-btn-secondary';
  document.getElementById('tab-file').className = tab === 'file' ? 'ctrl-btn-primary' : 'ctrl-btn-secondary';
}

function handleFileSelect(input) {
  if (input.files && input.files[0]) readSubFile(input.files[0]);
}

function handleFileDrop(e) {
  if (e.dataTransfer.files && e.dataTransfer.files[0]) readSubFile(e.dataTransfer.files[0]);
}

function readSubFile(file) {
  const reader = new FileReader();
  reader.onload = function(e) {
    subFileContent = e.target.result;
    document.getElementById('sub-file-label').innerHTML =
      '<span style="color:var(--fg)">✅ ' + file.name + '</span><br>' +
      '<span style="font-size:9px;opacity:0.6">' + (subFileContent.length / 1024).toFixed(1) + ' KB</span>';
  };
  reader.readAsText(file);
}

function openSubModal() {
  subFileContent = '';
  subTab = 'url';
  switchSubTab('url');
  document.getElementById('sub-modal').style.display = 'flex';
}

function closeSubModal() {
  document.getElementById('sub-modal').style.display = 'none';
}

async function addSubscription() {
  const name = document.getElementById('sub-name').value || t('sub.add_title');
  const url = document.getElementById('sub-url').value;
  const refreshMin = parseInt(document.getElementById('sub-refresh').value) || 60;

  const data = { name, refresh_min: refreshMin };

  if (subTab === 'url') {
    if (!url) { alert(t('msg.sub_url_required')); return; }
    data.url = url;
  } else {
    if (!subFileContent) { alert(t('msg.sub_file_required')); return; }
    data.file_content = subFileContent;
  }

  const result = await api('/api/subscription/add', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data)
  });

  if (result && result.error) {
    alert(t('msg.submit_failed') + result.error);
    return;
  }
  if (result && result.status === 'added') {
    closeSubModal();
    showToast(t('msg.sub_added'));
    document.getElementById('sub-name').value = '';
    document.getElementById('sub-url').value = '';
    subFileContent = '';
    document.getElementById('sub-file-label').innerHTML = '' + t('sub.file_drop') + '';
    setTimeout(loadSubscriptions, 3000);
    setTimeout(loadProxies, 5000);
  }
}

function openManualModal() {
  document.getElementById('manual-content').value = '';
  document.getElementById('manual-modal').style.display = 'flex';
}

function closeManualModal() {
  document.getElementById('manual-modal').style.display = 'none';
}

async function submitManualNodes() {
  const content = document.getElementById('manual-content').value;
  if (!content || !content.trim()) {
    alert(t('msg.manual_required'));
    return;
  }

  const btn = document.getElementById('manual-submit-btn');
  btn.textContent = t('proxy.add_manual_submitting');
  btn.disabled = true;

  const result = await api('/api/proxy/add-manual', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({content})
  });

  btn.textContent = t('proxy.add_manual_submit');
  btn.disabled = false;

  if (result && result.error) {
    alert(t('msg.submit_failed') + result.error);
    return;
  }

  if (result && result.status === 'added') {
    closeManualModal();
    showToast(t('msg.manual_added').replace('{0}', result.valid || 0).replace('{1}', result.invalid || 0));
    loadAll();
  }
}

async function refreshSub(id) {
  await api('/api/subscription/refresh', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({id: id})
  });
  showToast(t('msg.sub_refreshed'));
  setTimeout(loadSubscriptions, 3000);
}

async function refreshAllSubs() {
  await api('/api/subscription/refresh-all', {method: 'POST'});
  showToast(t('msg.sub_refresh_all'));
  setTimeout(loadSubscriptions, 3000);
}

async function toggleSub(id) {
  await api('/api/subscription/toggle', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({id: id})
  });
  loadSubscriptions();
}

async function deleteSub(id) {
  if (!confirm(t('msg.sub_delete_confirm'))) return;
  await api('/api/subscription/delete', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({id: id})
  });
  loadSubscriptions();
}

// ========== 访客贡献订阅 ==========

let contributeFileContent = '';
let contributeTab = 'url';

function switchContributeTab(tab) {
  contributeTab = tab;
  document.getElementById('contribute-url-group').style.display = tab === 'url' ? '' : 'none';
  document.getElementById('contribute-file-group').style.display = tab === 'file' ? '' : 'none';
  document.getElementById('ctab-url').className = tab === 'url' ? 'ctrl-btn-primary' : 'ctrl-btn-secondary';
  document.getElementById('ctab-file').className = tab === 'file' ? 'ctrl-btn-primary' : 'ctrl-btn-secondary';
}

function handleContributeFileSelect(input) {
  if (input.files && input.files[0]) readContributeFile(input.files[0]);
}
function handleContributeFileDrop(e) {
  if (e.dataTransfer.files && e.dataTransfer.files[0]) readContributeFile(e.dataTransfer.files[0]);
}
function readContributeFile(file) {
  const reader = new FileReader();
  reader.onload = function(e) {
    contributeFileContent = e.target.result;
    document.getElementById('contribute-file-label').innerHTML =
      '<span style="color:var(--fg)">✅ ' + file.name + '</span><br>' +
      '<span style="font-size:9px;opacity:0.6">' + (contributeFileContent.length / 1024).toFixed(1) + ' KB</span>';
  };
  reader.readAsText(file);
}

function openContributeModal() {
  contributeFileContent = '';
  contributeTab = 'url';
  switchContributeTab('url');
  document.getElementById('contribute-modal').style.display = 'flex';
}

function closeContributeModal() {
  document.getElementById('contribute-modal').style.display = 'none';
}

async function submitContribution() {
  const name = document.getElementById('contribute-name').value || t('contribute.title');
  const data = { name };

  if (contributeTab === 'url') {
    const url = document.getElementById('contribute-url').value;
    if (!url) { alert(t('msg.sub_url_required')); return; }
    data.url = url;
  } else {
    if (!contributeFileContent) { alert(t('msg.sub_file_required')); return; }
    data.file_content = contributeFileContent;
  }

  const btn = document.getElementById('contribute-submit-btn');
  btn.textContent = t('contribute.validating');
  btn.disabled = true;

  const result = await api('/api/subscription/contribute', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data)
  });

  btn.textContent = t('contribute.submit');
  btn.disabled = false;

  if (result && result.error) {
    alert(t('msg.submit_failed') + result.error);
    return;
  }
  if (result && result.status === 'contributed') {
    closeContributeModal();
    showToast(t('msg.contribute_thanks'));
    document.getElementById('contribute-name').value = '';
    document.getElementById('contribute-url').value = '';
    contributeFileContent = '';
    document.getElementById('contribute-file-label').innerHTML = '' + t('sub.file_drop') + '';
    setTimeout(loadSubscriptions, 3000);
  }
}

loadAll();
loadSubscriptions();
setInterval(loadPoolStatus, 5000);
setInterval(loadQualityDistribution, 10000);
setInterval(loadLogs, 5000);
setInterval(loadSubscriptions, 30000);

// 日志倒计时
setInterval(() => {
  logCountdown--;
  if (logCountdown < 0) logCountdown = 5;
  updateLogCountdown();
}, 1000);
</script>

<div id="toast" class="toast"></div>

</body>
</html>`
