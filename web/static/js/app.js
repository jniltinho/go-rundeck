/* Go-Rundeck — app.js */
$(document).ready(function () {

  // ── Toast helper ────────────────────────────────────────────────────────────
  window.toast = function (message, type) {
    type = type || 'info';
    var colors = {
      info:    'bg-blue-100 border-blue-500 text-blue-800',
      success: 'bg-green-100 border-green-600 text-green-800',
      error:   'bg-red-100 border-red-600 text-red-800',
      warn:    'bg-yellow-100 border-yellow-500 text-yellow-800',
    };
    var cls = colors[type] || colors.info;
    var $t = $('<div>')
      .addClass('fixed bottom-4 right-4 z-50 px-4 py-3 border-2 font-mono text-sm shadow-[4px_4px_0_#0C0C0C] ' + cls)
      .text(message);
    $('body').append($t);
    setTimeout(function () { $t.fadeOut(400, function () { $t.remove(); }); }, 3500);
  };

  // ── Confirm delete forms ─────────────────────────────────────────────────────
  $(document).on('submit', 'form[data-confirm]', function (e) {
    var msg = $(this).data('confirm') || 'Are you sure?';
    if (!window.confirm(msg)) {
      e.preventDefault();
      return false;
    }
  });

  // ── Auto-close flash messages ────────────────────────────────────────────────
  setTimeout(function () {
    $('.flash-message').fadeOut(400);
  }, 4000);

  // ── Navigation active state ──────────────────────────────────────────────────
  var path = window.location.pathname;
  $('aside nav a').each(function () {
    var href = $(this).attr('href');
    if (href && href !== '/' && path.startsWith(href)) {
      $(this).addClass('bg-[#FF5C00] text-black border-[#FF5C00]');
    } else if (href === '/' && path === '/') {
      $(this).addClass('bg-[#FF5C00] text-black border-[#FF5C00]');
    }
  });

});
