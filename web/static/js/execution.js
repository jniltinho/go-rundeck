/* Go-Rundeck — execution.js
 * SSE log streaming for execution detail page.
 */

/**
 * initExecLog sets up a Server-Sent Events connection to stream execution logs.
 *
 * @param {number} execID   - The execution ID
 * @param {string} status   - Current execution status ("running" | other)
 */
function initExecLog(execID, status) {
  var $output = $('#log-output');
  var $indicator = $('#log-status');

  if (status !== 'running') {
    $indicator.text('complete');
    scrollBottom($output);
    return;
  }

  if (!window.EventSource) {
    $indicator.text('SSE not supported');
    return;
  }

  $indicator.text('streaming…');

  var url = '/executions/' + execID + '/log';
  var es = new EventSource(url);

  es.onmessage = function (e) {
    appendLine($output, e.data, 'text-green-400');
    scrollBottom($output);
  };

  es.addEventListener('done', function (e) {
    $indicator.text('complete');
    appendLine($output, '--- execution complete ---', 'text-gray-500');
    scrollBottom($output);
    es.close();
  });

  es.onerror = function () {
    $indicator.text('disconnected');
    appendLine($output, '--- connection lost ---', 'text-red-400');
    es.close();
  };
}

function appendLine($container, text, cssClass) {
  cssClass = cssClass || 'text-green-400';
  var $line = $('<div>').addClass(cssClass).text(text);
  $container.append($line);
}

function scrollBottom($el) {
  $el.scrollTop($el[0].scrollHeight);
}
