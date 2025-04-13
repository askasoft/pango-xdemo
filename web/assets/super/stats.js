(function($) {
	function stats_build($d, data) {
		$d.find('.size').text(data.size);

		var i = 0;
		var $tb = $d.find('table > tbody').empty();
		$.each(data.data, function(k, v) {
			$tb.append($('<tr>').append(
				$('<td>').text(++i),
				$('<td>').text(v.key),
				$('<td>').text(v.val),
				$('<td>').text(v.ttl)
			));
		});
	}

	function stats_load($t, force) {
		if (!force && $t.data('loaded')) {
			return;
		}

		$.ajax({
			url: $t.attr('id').replace('aps_', ''),
			method: 'GET',
			beforeSend: $t.loadmask.delegate($t),
			success: function(data) {
				$t.data('loaded', true);
				if (typeof(data) == 'string') {
					$t.find('.stats').empty().text(data);
				} else {
					stats_build($t, data);
				}
			},
			error: main.ajax_error,
			complete: $t.unloadmask.delegate($t)
		})
	}

	function init() {
		$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function() {
			stats_load($($(this).attr('href')));
		});

		$('.s-stats button.reload').on('click', function() {
			stats_load($(this).closest('.aps'), true);
		});
	}

	$(window).on('load', init);
})(jQuery);

