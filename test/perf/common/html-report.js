// =============================================================================
// Standalone HTML Performance Report Generator for RealEstate-Trust
// Zero external dependencies - fully self-contained HTML/CSS
// =============================================================================

export function generateHtmlReport(data, scenarioName = 'Performance Test') {
  const metrics = data.metrics || {};
  const duration = metrics.http_req_duration ? metrics.http_req_duration.values : {};
  const reqs = metrics.http_reqs ? metrics.http_reqs.values : {};
  const failed = metrics.http_req_failed ? metrics.http_req_failed.values : {};
  const checks = metrics.checks ? metrics.checks.values : {};

  const totalReqs = reqs.count || 0;
  const rps = reqs.rate ? reqs.rate.toFixed(2) : '0';
  const p95 = duration['p(95)'] ? `${duration['p(95)'].toFixed(2)} ms` : 'N/A';
  const p99 = duration['p(99)'] ? `${duration['p(99)'].toFixed(2)} ms` : 'N/A';
  const avg = duration.avg ? `${duration.avg.toFixed(2)} ms` : 'N/A';
  const errorRate = failed.rate ? `${(failed.rate * 100).toFixed(2)}%` : '0.00%';
  const checkRate = checks.rate ? `${(checks.rate * 100).toFixed(1)}%` : '100%';

  const isSuccess = (!failed.rate || failed.rate < 0.05) && (!checks.rate || checks.rate > 0.90);
  const statusBadge = isSuccess
    ? '<span style="background:#10B981; color:#fff; padding:4px 12px; border-radius:12px; font-weight:bold;">PASSED</span>'
    : '<span style="background:#EF4444; color:#fff; padding:4px 12px; border-radius:12px; font-weight:bold;">FAILED</span>';

  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>RealEstate-Trust Performance Report - ${scenarioName}</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0F172A; color: #F8FAFC; margin: 0; padding: 24px; }
    .container { max-width: 1000px; margin: 0 auto; }
    .header { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #334155; padding-bottom: 16px; margin-bottom: 24px; }
    .title { font-size: 24px; font-weight: 700; color: #38BDF8; }
    .subtitle { color: #94A3B8; font-size: 14px; margin-top: 4px; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 16px; margin-bottom: 28px; }
    .card { background: #1E293B; border: 1px solid #334155; border-radius: 8px; padding: 16px; text-align: center; }
    .card-label { font-size: 12px; text-transform: uppercase; color: #94A3B8; letter-spacing: 0.05em; }
    .card-val { font-size: 22px; font-weight: 700; color: #F8FAFC; margin-top: 6px; }
    .table-container { background: #1E293B; border: 1px solid #334155; border-radius: 8px; padding: 20px; margin-bottom: 24px; }
    table { width: 100%; border-collapse: collapse; margin-top: 12px; font-size: 14px; }
    th { text-align: left; color: #94A3B8; padding: 8px 12px; border-bottom: 1px solid #334155; }
    td { padding: 10px 12px; border-bottom: 1px solid #1E293B; color: #E2E8F0; }
    tr:nth-child(even) td { background: #182234; }
    .footer { text-align: center; color: #64748B; font-size: 12px; margin-top: 32px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <div>
        <div class="title">🏢 RealEstate-Trust Performance Report</div>
        <div class="subtitle">Scenario: ${scenarioName} &bull; Generated: ${new Date().toISOString()}</div>
      </div>
      <div>${statusBadge}</div>
    </div>

    <div class="grid">
      <div class="card">
        <div class="card-label">Total Requests</div>
        <div class="card-val">${totalReqs}</div>
      </div>
      <div class="card">
        <div class="card-label">Throughput</div>
        <div class="card-val">${rps} req/s</div>
      </div>
      <div class="card">
        <div class="card-label">p(95) Latency</div>
        <div class="card-val">${p95}</div>
      </div>
      <div class="card">
        <div class="card-label">p(99) Latency</div>
        <div class="card-val">${p99}</div>
      </div>
      <div class="card">
        <div class="card-label">Error Rate</div>
        <div class="card-val" style="color:${failed.rate > 0.02 ? '#EF4444' : '#10B981'};">${errorRate}</div>
      </div>
      <div class="card">
        <div class="card-label">Checks Passed</div>
        <div class="card-val" style="color:#38BDF8;">${checkRate}</div>
      </div>
    </div>

    <div class="table-container">
      <h3 style="margin-top:0; color:#38BDF8;">Latency Distribution</h3>
      <table>
        <thead>
          <tr>
            <th>Metric</th>
            <th>Min</th>
            <th>Average</th>
            <th>Median</th>
            <th>p(90)</th>
            <th>p(95)</th>
            <th>p(99)</th>
            <th>Max</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>HTTP Request Duration</td>
            <td>${duration.min ? duration.min.toFixed(2) + ' ms' : '-'}</td>
            <td>${avg}</td>
            <td>${duration.med ? duration.med.toFixed(2) + ' ms' : '-'}</td>
            <td>${duration['p(90)'] ? duration['p(90)'].toFixed(2) + ' ms' : '-'}</td>
            <td>${p95}</td>
            <td>${p99}</td>
            <td>${duration.max ? duration.max.toFixed(2) + ' ms' : '-'}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="footer">
      RealEstate-Trust Monorepo Platform &bull; Automated Institutional Benchmark
    </div>
  </div>
</body>
</html>`;
}
