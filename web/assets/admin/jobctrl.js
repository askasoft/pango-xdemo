$(function() {
	function job_start() {
		if (!$('#job_start').prop('disabled')) {
			$('#job_start').prop('disabled', true);

			var $jf = $('#job_form'), jid;
			$.ajaf({
				url: './start',
				type: 'POST',
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
		}
		return false;
	}

	function job_abort() {
		$(this).prop('disabled', true).find('i').addClass('fa-spinner fa-spin');

		var jid = $(this).closest('.job').data('jid');

		$.ajax({
			url: './abort',
			type: 'POST',
			data: {
				_token_: main.token,
				jid: jid
			},
			dataType: 'json',
			success: main.ajax_success,
			error: main.ajax_error,
			complete: function() {
				job_list(jid);
			}
		});
		return false;
	}

	function job_status(jid) {
		if ($('#job_head_' + jid).find('a.active').length == 0) {
			return;
		}

		var $job = $('#job_' + jid);

		if ($job.data('timer')) {
			clearTimeout($job.data('timer'));
			$job.data('timer', null);
		}

		var timeout = 0, param = { jid: jid, asc: false, limit: 1000 };
		var $tb = $job.find('tbody'), $tr = $tb.children('tr:first-child');

		if ($tr.length) {
			param.min = parseInt($tr.data('lid')) + 1;
			param.asc = true;
		}

		$.ajax({
			url: './status',
			data: param,
			type: 'GET',
			dataType: 'json',
			success: function(data) {
				var job = data.job, logs = data.logs || [];

				job_info_refresh(job);
				job_start_refresh();

				if (logs.length >= param.limit) {
					// has next logs
					timeout = 100;
				} else if (job.status == 'P' || job.status == 'R' || new Date().getTime() - new Date(job.updated_at).getTime() < 10000) {
					// running, pending or updated_at is in 10sec
					timeout = 1000;
				}
				if (timeout) {
					$job.data('timer', setTimeout(function() { job_status(jid); }, timeout));
				}

				if (logs.length > 0) {
					var tb = $tb.get(0), c = document.createDocumentFragment();;
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
			type: 'GET',
			dataType: 'json',
			beforeSend: function() { $a.hide(); $i.show(); },
			success: function(logs) {
				if (logs && logs.length > 0) {
					var tb = $tb.get(0), c = document.createDocumentFragment();;
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

	function job_start_refresh() {
		if ($('#job_form').data('multi')) {
			$('#job_start').prop('disabled', false);
		} else {
			var $jhs = $('#job_list > ul'), jpr = $jhs.find('li.P, li.R').length > 0;
			$('#job_start').prop('disabled', jpr).find('i')[jpr ? 'addClass' : 'removeClass']('fa-spinner fa-spin');
		}
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

		var $jl = $('#job_list'), $jhs = $jl.children('ul');
		if ($jhs.length == 0) {
			$jhs = $('<ul class="nav nav-pills">');
			$jhs.on('click', 'a', job_tab_show);
			$jhs.on('shown.bs.tab', 'a', job_tab_shown);
			$jl.append($jhs);
		}

		var $jobs = $jl.children('div');
		if ($jobs.length == 0) {
			$jobs = $('<div class="tab-content">');
			$jl.append($jobs);
		}

		for (var i = data.length - 1; i >= 0; i--) {
			var job = data[i], $jh = $jhs.children('#job_head_' + job.id);
			if ($jh.length) {
				job_info_refresh(job);
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

		$jhs.find(jid ? '#job_head_' + jid + ' > a' : 'a:first').trigger('click');
	}

	function build_job_head(job) {
		var $jh = $('<li>', { id: 'job_head_' + job.id, 'class': 'nav-item' }).attr('status', job.status);
		$jh.data('jid', job.id);

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
		var $jf = $('#job_form'), legend = $jf.prev('legend').text();
		if (!legend || !(job.file || job.param)) {
			return $('<hr/>');
		}

		var $fset = $('<fieldset>', { 'class': 'ui-fieldset collapsed' }).append($('<legend>').text(legend));

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
				if (typeof(v) == 'string') {
					var d = new Date(v);
					if (d.getTime() > 0 ) {
						params[k] = main.format_date(d);
					}
				}
			}
			$cf.formValues(params);

			var copy = $jf.data('copy');
			if (copy) {
				var $a = $('<a>', { 'class': 'btn btn-secondary ps', href: '#' });
				$a.append($('<i class="fas fa-arrow-up">'));
				$a.append($('<span>').text(copy));
				$a.on('click', job_copy_param);
				$cf.append($a);
			}
		}

		$cf.find('select[data-spy="niceSelect"]').niceSelect();

		$fset.append($cf).fieldset().on('expanded.fieldset', function() {
			$(this).find('textarea').autosize();
		});
		return $fset;
	}

	function build_job_tools(jst) {
		var $jt = $('<div>', { 'class': 'job-tools' });

		if (jst == 'P' || jst == 'R') {
			$jt.append(build_job_abort());
		}
		return $jt;
	}

	function build_job_abort() {
		var $jf = $('#job_form');

		var $btn = $('<button>', { 'class': 'abort btn btn-danger' });
		var $i = $('<i>', { 'class': 'fas fa-stop' });
		var $t = $('<span>').text($jf.data('abort'));

		$btn.append($i, $t);
		return $btn;
	}

	function job_info_refresh(job) {
		var $jh = $('#job_head_' + job.id);
		if ($jh.attr('status') != job.status) {
			$jh.attr('status', job.status);
			$jh.find('i').attr('class', job_status_icon(job.status));

			var $jt = $('#job_' + job.id).find('.job-tools').empty();
			if (job.status == 'P' || job.status == 'R') {
				$jt.append(build_job_abort());
			}
		}

		build_job_abort($jt, job.status);
	}

	function job_list(jid) {
		var $jhs = $('#job_list > ul');

		$.ajax({
			url: './list',
			type: 'GET',
			dataType: 'json',
			beforeSend: $jhs.loadmask.delegate($jhs),
			success: function(data) {
				setTimeout(function() {
					build_job_list(data, jid);
					job_start_refresh();
				}, 10);
			},
			error: main.ajax_error,
			complete: $jhs.unloadmask.delegate($jhs)
		});

		return false;
	}

	$('#job_form').on('submit', job_start);
	$('#job_start').on('click', job_start);
	$('#job_list').on('click', 'button.abort', job_abort);

	job_list();
});
