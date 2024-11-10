(function($) {
	var jobnames = {}, jslabels = {};

	function jobchain_start() {
		$('#jobchain_start').prop('disabled', true);

		var $jf = $('#jobchain_form'), cid;
		$.ajaf({
			url: './start',
			method: 'POST',
			data: $jf.serializeArray(),
			file: $jf.find('input[type="file"]'),
			dataType: 'json',
			beforeSend: main.form_ajax_start($jf),
			success: function(data) {
				cid = data.cid;
				main.ajax_success(data);
			},
			error: main.form_ajax_error($jf),
			complete: function() {
				main.form_ajax_end($jf)();
				jobchain_list(cid);
			}
		});
		return false;
	}

	function jobchain_abort() {
		var $i = $(this).prop('disabled', true).find('i').addClass('fa-spinner fa-spin');

		var cid = $(this).closest('.jobchain').data('cid');
		if (!cid) {
			cid = $('#jobchain_tabs > ul').children('li[status=P], li[status=R]').data('cid');
		}

		$.ajax({
			url: './abort',
			method: 'POST',
			data: {
				cid: cid
			},
			dataType: 'json',
			success: main.ajax_success,
			error: main.ajax_error,
			complete: function() {
				$i.removeClass('fa-spinner fa-spin');
				jobchain_status(cid);
			}
		});
		return false;
	}

	function jobchain_status(cid) {
		var $jch = $('#jobchain_head_' + cid);
		if ($jch.find('a.active').length == 0 || $jch.hasClass('done')) {
			return;
		}

		if ($jch.data('timer')) {
			clearTimeout($jch.data('timer'));
			$jch.data('timer', null);
		}

		$.ajax({
			url: './status',
			data: { cid: cid },
			method: 'GET',
			dataType: 'json',
			success: function(data) {
				refresh_jobchain_info(data);
				refresh_jobchain_start();

				if (jobchain_is_done(data)) {
					$jch.addClass('done');
				} else {
					$jch.data('timer', setTimeout(function() { jobchain_status(cid); }, 1000));
				}
			},
			error: main.ajax_error
		});
	}

	function jobchain_is_done(jci) {
		// (completed or aborted) and (updated_at is after 10sec)
		return (jci.status == 'C' || jci.status == 'A') && (new Date().getTime() - new Date(jci.updated_at).getTime() > 10000);
	}

	function refresh_jobchain_start() {
		if ($('#jobchain_form').data('multi')) {
			$('#jobchain_start').prop('disabled', false);
			return;
		}

		var jpr = $('#jobchain_tabs > ul').children('li[status=P], li[status=R]').length > 0;
		var $b = $('#jobchain_start').prop('disabled', jpr), lbl = $b.data('processing');
		if (lbl) {
			var $s = $b.find('span');
			if (jpr) {
				if (!$b.data('original')) {
					$b.data('original', $s.text());
				}
				$s.text(lbl);
			} else {
				$s.text($b.data('original'));
			}
		}
		$b.find('i')[jpr ? 'addClass' : 'removeClass']('fa-spinner fa-spin');

		$('#jobchain_abort').prop('disabled', !jpr);
	}

	function jobchain_status_icon(s) {
		switch (s) {
		case 'A':
			return 'far fa-circle-xmark';
		case 'C':
			return 'far fa-circle-check';
		case 'P':
			return 'fas fa-circle-notch fa-spin';
		case 'R':
			return 'fas fa-rotate fa-spin';
		default:
			return '';
		}
	}

	function jobchain_tab_show() {
		$(this).tab('show');
		return false;
	}

	function jobchain_tab_shown(e) {
		jobchain_status($(e.target).data('cid'));
		return false;
	}

	function build_jobchain_list(data, cid) {
		if (!data || data.length == 0) {
			return;
		}

		var $jchs = $('#jobchain_tabs > ul');
		var $jcs = $('#jobchain_list');

		for (var i = data.length - 1; i >= 0; i--) {
			var jci = data[i], $jch = $jchs.children('#jobchain_head_' + jci.id);
			if ($jch.length) {
				refresh_jobchain_info(jci);
				continue;
			}

			$jch = build_jobchain_head(jci);
			$jchs.prepend($jch);

			var $jci = build_jobchain_info(jci);
			$jcs.prepend($jci);

			if (jobchain_is_done(jci)) {
				$jch.addClass('done');
			}
		}

		$jchs.find(cid ? '#jobchain_head_' + cid + ' > a' : 'li:first-child > a').trigger('click');
	}

	function build_jobchain_head(jci) {
		var $jch = $('<li>', { id: 'jobchain_head_' + jci.id, 'class': 'nav-item' }).attr('status', jci.status).data('cid', jci.id);

		var $a = $('<a>', { href: '#jobchain_' + jci.id, 'class': 'nav-link' });
		$a.data('cid', jci.id);
		$a.append($('<i>', { 'class': jobchain_status_icon(jci.status) }));
		$a.append($('<span>').text(main.format_time(jci.created_at)));

		$jch.append($a);
		return $jch;
	}

	function build_jobchain_info(jci) {
		var $jci = $('<div>', { id: 'jobchain_' + jci.id, 'class': 'jobchain tab-pane fade' }).attr('status', jci.status).data('cid', jci.id);

		$jci.append(build_jobchain_tools(jci));

		var $jss = $('<div>', { 'class': 'jobstates'} );
		var $tpl = $('#jobchain_template > .jrs');

		$.each(jci.states, function(i, jrs) {
			var $jrs = $tpl.clone();
			jobchain_jrs_refresh($jrs, jrs);
			$jss.append($jrs);
		});

		$jci.append($jss);
		return $jci;
	}

	function build_jobchain_tools(jci) {
		var $jt = $('<div>', { 'class': 'jobtools' });

		append_jobchain_abort($jt, jci.status);
		append_jobchain_caption($jt, jci);
		return $jt;
	}

	function append_jobchain_caption($jt, jci) {
		var $jcp = $('<div>', { 'class': 'jobcaption' });
		$jcp.append($('<i>', { 'class': jobchain_status_icon(jci.status) }));
		$jcp.append($('<span>').text(jci.caption));
		$jt.append($jcp);
	}

	function append_jobchain_abort($jt, jst) {
		if (jst == 'P' || jst == 'R') {
			var label = $('#jobchain_form').data('abort');
			if (label) {
				var $b = $('<button>', { 'class': 'abort btn btn-danger' }),
					$i = $('<i>', { 'class': 'fas fa-stop' }),
					$t = $('<span>').text(label);
				$jt.append($b.append($i, $t));
			}
		}
	}

	function refresh_jobchain_info(jci) {
		var $jc = $('#jobchain_' + jci.id);
		var $jch = $('#jobchain_head_' + jci.id);

		if ($jch.attr('status') != jci.status) {
			$jch.attr('status', jci.status);
			$jch.find('i').attr('class', jobchain_status_icon(jci.status));

			$jc.attr('status', jci.status);

			var $jt = $jc.find('.jobtools').empty();
			append_jobchain_abort($jt, jci.status);
			append_jobchain_caption($jt, jci);
		}

		var $jrss = $jc.find('.jrs');
		$.each(jci.states, function(i, jrs) {
			jobchain_jrs_refresh($jrss.eq(i), jrs);
		});
	}

	function jobchain_jrs_set_state($jrs, jrs, key, val) {
		var $d = $jrs.find('div.' + key);
		$d.find('label').text(jslabels[jrs.name + '.' + key] || jslabels[key] || key);

		val ||= jrs.state[key];
		$d.find('span').text(Number.comma(val));
		$d[val ? 'show' : 'hide']();
	}

	function jobchain_jrs_set_error($jrs, jrs, key) {
		var $d = $jrs.find('div.' + key).data(key, jrs[key]);
		$d.find('label').text(jslabels[jrs.name + '.' + key] || jslabels[key] || key);
		$d[jrs[key] ? 'show' : 'hide']();
	}

	function jobchain_jrs_refresh($jrs, jrs) {
		$jrs.attr('jid', jrs.jid);
		$jrs.attr('status', jrs.status);
		$jrs.find('.jnm').text(jobnames[jrs.name] || jrs.name);

		jobchain_jrs_set_state($jrs, jrs, 'total', jrs.state.limit || jrs.state.total || '-');
		jobchain_jrs_set_state($jrs, jrs, 'skipped');
		jobchain_jrs_set_state($jrs, jrs, 'success', jrs.state.success || '-');
		jobchain_jrs_set_state($jrs, jrs, 'failure');
		jobchain_jrs_set_error($jrs, jrs, 'error');

		var p = 0, t = '・・・';
		if (jrs.status == 'C') {
			p = 100;
			t = '100%';
		} else if (jrs.state.limit) {
			p = Math.floor((jrs.state.success || 0) * 100 / jrs.state.limit)
			t = p + '%';
		} else if (jrs.state.total) {
			p = Math.floor((jrs.state.step || 0) * 100 / jrs.state.total)
			t = p + '%';
		} else {
			p = 10;
		}
		$jrs.find('.txt').text(t);

		var size = 80, stroke_width = 8;
		var radius = (size - stroke_width) / 2;
		var circumference = radius * 3.14 * 2;
		var dash = p * circumference / 100;
		var stroke_dasharray = dash + ' ' + (circumference - dash);
		$jrs.find('svg .fg').css('stroke-dasharray', stroke_dasharray);
	}

	function jobchain_list(cid) {
		var $jchs = $('#jobchain_tabs > ul');

		$.ajax({
			url: './list',
			method: 'GET',
			dataType: 'json',
			beforeSend: $jchs.loadmask.bind($jchs),
			success: function(data) {
				setTimeout(function() {
					if (data && data.length && !cid) {
						$('#jobchain_history').fieldset('expand');
					}
					build_jobchain_list(data, cid);
					refresh_jobchain_start();
				}, 10);
			},
			error: main.ajax_error,
			complete: $jchs.unloadmask.bind($jchs)
		});

		return false;
	}

	function jobchain_error() {
		$('#jobchain_error_popup .ui-popup-body').empty().text($(this).data('error'));
		$('#jobchain_error_popup').popup('toggle', this);
		return false;
	}

	function jobchain_init() {
		jobnames = $('#jobchain_maps').data('jobnames');
		jslabels = $('#jobchain_maps').data('jslabels');

		$('#jobchain_form').on('submit', jobchain_start);
		$('#jobchain_start').on('click', jobchain_start);
		$('#jobchain_abort').on('click', jobchain_abort);

		$('#jobchain_tabs')
			.on('click', 'a', jobchain_tab_show)
			.on('shown.bs.tab', 'a', jobchain_tab_shown);

		$('#jobchain_list')
			.on('click', 'button.abort', jobchain_abort)
			.on('click', 'div.error', jobchain_error);

		jobchain_list();
	}

	$(window).on('load', jobchain_init);
})(jQuery);
