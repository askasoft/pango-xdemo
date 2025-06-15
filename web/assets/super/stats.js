(function($) {
	function stats_cache_build($d, data) {
		$d.find('.size').text(data.size);

		var i = 0, $tb = $d.find('table > tbody').empty();

		$.each(data.data, function(k, v) {
			$tb.append($('<tr>').append(
				$('<td>').text(++i),
				$('<td>').text(v.key),
				$('<td>').text(v.val),
				$('<td>').text(v.ttl)
			));
		});
	}

	function stats_items_build($d, data) {
		var i = 0, $tb = $d.find('table > tbody').empty();

		for (var k in data) {
			$tb.append($('<tr>').append(
				$('<td>').text(++i),
				$('<td>').text(k),
				$('<td>').text(data[k])
			));
		}
	}

	function stats_build($d, data) {
		if (typeof(data) == 'string') {
			$t.find('.jobstats').empty().text(data);
		} else if (data.type == 'cache') {
			stats_cache_build($d, data);
		} else {
			stats_items_build($d, data);
		}
	}

	function stats_load($t, force) {
		if (!force && $t.data('loaded')) {
			return;
		}

		$.ajax({
			url: $t.attr('id').replace('aps_', '').replace('_', '/'),
			method: 'GET',
			beforeSend: $t.loadmask.delegate($t),
			success: function(data) {
				$t.data('loaded', true);
				stats_build($t, data);
			},
			error: main.ajax_error,
			complete: $t.unloadmask.delegate($t)
		})
	}

	function init() {
		$('.s-stats > ul.nav a[data-bs-toggle="tab"]').on('shown.bs.tab', function() {
			var t = $(this).attr('href');
			if (t == '#aps_cache') {
				stats_load($($('#aps_cache ul.nav a.active').attr('href')));
			}
			history.replaceState(null, null, location.href.split('#')[0] + t);
		});

		$('#aps_cache a[data-bs-toggle="tab"]').on('shown.bs.tab', function() {
			stats_load($($(this).attr('href')));
		});

		$('.s-stats button.reload').on('click', function() {
			stats_load($(this).closest('.aps'), true);
		});

		var $t = $('a[href="' + location.hash + '"]');
		if ($t.length) {
			new bootstrap.Tab($t.get(0)).show();
		}
	}

	$(window).on('load', init);
})(jQuery);

