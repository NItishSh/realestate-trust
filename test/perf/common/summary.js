// =============================================================================
// Reusable k6 handleSummary Processor
// Exports latest.json and interactive latest.html report
// =============================================================================

import { generateHtmlReport } from './html-report.js';

export function createSummaryHandler(scenarioName = 'Test Scenario') {
  return function (data) {
    data.scenarioName = scenarioName;
    return {
      'reports/latest.json': JSON.stringify(data, null, 2),
      'reports/latest.html': generateHtmlReport(data, scenarioName),
    };
  };
}
