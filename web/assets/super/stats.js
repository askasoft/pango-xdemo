(function($) {
	function stats_cache_build($d, data) {
		$d.find('.total').text(data.total);

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

	function stats_cards_build($d, data) {
		var $cs = $d.find('.cards').empty();

		for (var k in data) {
			$cs.append($('<div class="card">').append(
				$('<div class="card-header">').text(k),
				$('<div class="card-body">').text(data[k])
			));
		}
	}

	function stats_build($d, uri, data) {
		switch (uri) {
		case 'db':
		case 'server':
			stats_cards_build($d, data);
			break;
		case 'jobs':
			$d.find('pre').empty().text(data);
			break;
		default:
			stats_cache_build($d, data);
			break;
		}
	}

	function stats_load($t, force) {
		if (!force && $t.data('loaded')) {
			return;
		}

		var uri = $t.attr('id').replace('stats_', '').replace('_', '/');

		$.ajax({
			url: uri,
			method: 'GET',
			beforeSend: $t.loadmask.delegate($t),
			success: function(data) {
				$t.data('loaded', true);
				stats_build($t, uri, data);
			},
			error: main.ajax_error,
			complete: $t.unloadmask.delegate($t)
		})
	}

	function init() {
		$('.s-stats > ul.nav a[data-bs-toggle="tab"]').on('shown.bs.tab', function() {
			var t = $(this).attr('href');
			if (t == '#stats_cache') {
				stats_load($($('#stats_cache ul.nav a.active').attr('href')));
			} else {
				stats_load($(t));
			}
			history.replaceState(null, null, location.href.split('#')[0] + t);
		});

		$('#stats_cache a[data-bs-toggle="tab"]').on('shown.bs.tab', function() {
			stats_load($($(this).attr('href')));
		});

		$('.s-stats button.reload').on('click', function() {
			stats_load($(this).closest('.stats'), true);
		});

		var $t = $('a[href="' + location.hash + '"]');
		if ($t.length) {
			new bootstrap.Tab($t.get(0)).show();
		}
	}

	$(window).on('load', init);
})(jQuery);

