(function($) {
	function job_start() {
		$('#job_start').prop('disabled', true);

		var $jf = $('#job_form'), jid;
		$.ajaf({
			url: './start',
			method: 'POST',
			data: $jf.serializeArray(),
			file: $jf.find('input[type="file"]'),
			dataType: 'json',
			beforeSend: main.form_ajax_start($jf),
			success: function(data) {
				jid = data.jid;
				main.ajax_success(data);
			},
			error: main.form_ajax_error($jf),
			complete: function() {
				main.form_ajax_end($jf)();
				job_list(jid);
			}
		});
		return false;
	}

	function job_abort() {
		var $i = $(this).prop('disabled', true).find('i').addClass('fa-spinner fa-spin');

		var jid = $(this).closest('.job').data('jid');
		if (!jid) {
			jid = $('#job_tabs > ul').children('li[status=P], li[status=R]').data('jid');
		}

		$.ajax({
			url: './abort',
			method: 'POST',
			data: {
				jid: jid
			},
			dataType: 'json',
			success: main.ajax_success,
			error: main.ajax_error,
			complete: function() {
				$i.removeClass('fa-spinner fa-spin');
				job_list(jid);
			}
		});
		return false;
	}

	function job_status(jid) {
		var $jh = $('#job_head_' + jid);
		if ($jh.find('a.active').length == 0 || $jh.hasClass('done')) {
			return;
		}

		if ($jh.data('timer')) {
			clearTimeout($jh.data('timer'));
			$jh.data('timer', null);
		}

		var param = { jid: jid, asc: false, limit: 1000 };

		var $tb = $('#job_' + jid).find('tbody'), $tr = $tb.children('tr:first-child');
		if ($tr.length) {
			param.min = parseInt($tr.data('lid')) + 1;
			param.asc = true;
		}

		$.ajax({
			url: './status',
			data: param,
			method: 'GET',
			dataType: 'json',
			success: function(data) {
				var job = data.job, logs = data.logs || [];

				refresh_job_info(job);
				refresh_job_start();

				var timeout;
				if (logs.length >= param.limit) {
					// has next logs
					timeout = 100;
				} else if (job.status == 'P' || job.status == 'R' || new Date().getTime() - new Date(job.updated_at).getTime() < 10000) {
					// running, pending or updated_at is in 10sec
					timeout = 1000;
				}
				if (timeout) {
					$jh.data('timer', setTimeout(function() { job_status(jid); }, timeout));
				} else {
					$jh.addClass('done');
				}

				if (logs.length > 0) {
					var tb = $tb.get(0), c = document.createDocumentFragment();
					if (param.asc) {
						for (var i = logs.length - 1; i >= 0; i--) {
							c.append(build_log(logs[i]));
						}
					} else {
						for (var i = 0; i < logs.length; i++) {
							c.append(build_log(logs[i]));
						}
					}
					tb.prepend(c);

					if (!param.asc && logs.length >= param.limit) {
						// first time to get latest logs
						var $a = $('<a>', { href: '#' }).text('...').on('click', job_prev_logs);
						var $i = $('<i>', { 'class': 'fas fa-spinner fa-spin-pulse fa-fw' }).hide();
						$tb.append($('<tr>').append($('<td>'), $('<td>'), $('<td>').append($a, $i)));
					}
				}
			},
			error: main.ajax_error
		});
	}

	function job_prev_logs() {
		var $a = $(this), $i = $a.next();
		var $td = $a.closest('td'), $tr = $td.closest('tr'), $tb = $tr.closest('tbody');
		var jid = $tb.closest('.job').data('jid');
		var lid = parseInt($tr.prev().data('lid'));
		var param = { jid: jid, asc: false, max: lid - 1, limit: 10000 };

		$.ajax({
			url: './logs',
			data: param,
			method: 'GET',
			dataType: 'json',
			beforeSend: function() { $a.hide(); $i.show(); },
			success: function(logs) {
				if (logs && logs.length > 0) {
					var tb = $tb.get(0), c = document.createDocumentFragment();
					for (var i = 0; i < logs.length; i++) {
						c.append(build_log(logs[i]));
					}
					tb.append(c);
				}

				if (logs && logs.length >= param.limit) {
					$i.hide(); $a.show();
					$tb.append($tr);
				} else {
					$tr.remove();
				}
			},
			error: main.ajax_error
		});
		return false;
	}

	function build_log(lg) {
		var tr = document.createElement('tr'),
			td1 = document.createElement('td'),
			td2 = document.createElement('td'),
			td3 = document.createElement('td');
		tr.className = lg.level;
		tr.setAttribute('data-lid', lg.id);
		td1.textContent = main.format_time(lg.time);
		td2.textContent = '[' + lg.level + ']';
		td3.textContent = lg.message;
		tr.append(td1, td2, td3);
		return tr;
	}

	function refresh_job_start() {
		if ($('#job_form').data('multi')) {
			$('#job_start').prop('disabled', false);
			return;
		}

		var jpr = $('#job_tabs > ul').children('li[status=P], li[status=R]').length > 0;
		var $b = $('#job_start').prop('disabled', jpr), lbl = $b.data('processing');
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

		$('#job_abort').prop('disabled', !jpr);
	}

	function job_status_icon(s) {
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

	function job_tab_show() {
		$(this).tab('show');
		return false;
	}

	function job_tab_shown(e) {
		job_status($(e.target).data('jid'));
		return false;
	}

	function build_job_list(data, jid) {
		if (!data || data.length == 0) {
			return;
		}

		var $jhs = $('#job_tabs > ul');
		var $jobs = $('#job_list');

		for (var i = data.length - 1; i >= 0; i--) {
			var job = data[i], $jh = $jhs.children('#job_head_' + job.id);
			if ($jh.length) {
				refresh_job_info(job);
				continue;
			}

			$jh = build_job_head(job);
			$jhs.prepend($jh);

			var $job = $('<div>', { id: 'job_' + job.id, 'class': 'job tab-pane fade' }).data('jid', job.id);

			$job.append(build_job_param(job));
			$job.append(build_job_tools(job.status));

			$job.append($('<table>', { 'class': 'table table-striped' }).append($('<tbody>')));

			$jobs.prepend($job);
		}

		$jhs.find(jid ? '#job_head_' + jid + ' > a' : 'li:first-child > a').trigger('click');
	}

	function build_job_head(job) {
		var $jh = $('<li>', { id: 'job_head_' + job.id, 'class': 'nav-item' }).attr('status', job.status).data('jid', job.id);

		var $a = $('<a>', { href: '#job_' + job.id, 'class': 'nav-link' });
		$a.data('jid', job.id);
		$a.append($('<i>', { 'class': job_status_icon(job.status) }));
		$a.append($('<span>').text(main.format_time(job.created_at)));

		$jh.append($a);
		return $jh;
	}

	function job_copy_param() {
		var $cf = $(this).closest('form');

		$cf.find(':input').prop('disabled', false);
		var vs = $cf.formValues();
		$cf.find(':input').prop('disabled', true);

		$('#job_form').formValues(vs, true);
		return false;
	}

	function build_job_param(job) {
		var $jf = $('#job_form');
		if (!(job.file || job.param) || !$jf.data('form')) {
			return $('<hr/>');
		}

		var $fset = $('<fieldset>', { 'class': 'ui-fieldset' + (!$jf.data('expand') ? ' collapsed' : '') });
		
		$fset.append($('<legend>').text($jf.data('form')));

		var $cf = $jf.clone().attr('id', 'job_param_' + job.id);
		$cf.on('submit', function() { return false; });
		$cf.find('[type=hidden]').remove();
		$cf.find(':input').prop('disabled', true);
		$cf.formClear();

		var $f = $cf.find('[type=file]').hide();
		if (job.file) {
			$('<a>', { 'class': 'btn btn-secondary', href: main.base + '/files' + job.file })
				.append($('<i>', { 'class': 'fas fa-download' }))
				.insertAfter($f);
		}

		if (job.param) {
			var params = main.safe_parse_json(job.param);
			for (var k in params) {
				var v = params[k];
				switch (typeof(v)) {
				case 'string':
					var d = new Date(v);
					if (d.getTime() > 0 ) {
						params[k] = main.format_date(d);
					}
					break;
				case 'boolean':
					params[k] = v + '';
					break;
				}
			}
			$cf.formValues(params);

			var copy = $jf.data('copy');
			if (copy) {
				var $b = $('<button>', { 'class': 'btn btn-secondary copy' });
				$b.append($('<i class="fas fa-arrow-up">'));
				$b.append($('<span>').text(copy));
				$b.on('click', job_copy_param);
				$cf.append($b);
			}
		}

		$cf.find('select[data-spy="niceSelect"]').niceSelect();

		$fset.append($cf).fieldset().on('expanded.fieldset', function() {
			$(this).find('textarea').autosize();
		});
		return $fset;
	}

	function build_job_tools(jst) {
		var $jt = $('<div>', { 'class': 'jobtools' });
		append_job_abort($jt, jst);
		return $jt;
	}

	function append_job_abort($jt, jst) {
		if (jst == 'P' || jst == 'R') {
			var label = $('#job_form').data('abort');
			if (label) {
				var $btn = $('<button>', { 'class': 'abort btn btn-danger' }),
					$i = $('<i>', { 'class': 'fas fa-stop' }),
					$t = $('<span>').text(label);
				$jt.append($btn.append($i, $t));
			}
		}
	}

	function refresh_job_info(job) {
		var $jh = $('#job_head_' + job.id);
		if ($jh.attr('status') != job.status) {
			$jh.attr('status', job.status);
			$jh.find('i').attr('class', job_status_icon(job.status));

			var $jt = $('#job_' + job.id).find('.jobtools').empty();
			append_job_abort($jt, job.status);
		}
	}

	function job_list(jid) {
		var $jhs = $('#job_tabs > ul');

		$.ajax({
			url: './list',
			method: 'GET',
			dataType: 'json',
			beforeSend: $jhs.loadmask.bind($jhs),
			success: function(data) {
				setTimeout(function() {
					if (data && data.length && !jid) {
						$('#job_history').fieldset('expand');
					}
					build_job_list(data, jid);
					refresh_job_start();
				}, 10);
			},
			error: main.ajax_error,
			complete: $jhs.unloadmask.bind($jhs)
		});

		return false;
	}

	function job_init() {
		$('#job_form').on('submit', job_start);
		$('#job_start').on('click', job_start);
		$('#job_abort').on('click', job_abort);

		$('#job_tabs')
			.on('click', 'a', job_tab_show)
			.on('shown.bs.tab', 'a', job_tab_shown);

		$('#job_list').on('click', 'button.abort', job_abort);

		job_list();
	}

	$(window).on('load', job_init);
})(jQuery);
