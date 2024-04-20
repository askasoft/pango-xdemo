$(function() {
	function job_start() {
		$('#job_start').prop('disabled', true);

		var $f = $('#job_form');
		$.ajaf({
			url: './start',
			type: 'POST',
			data: $f.serializeArray(),
			file: $f.find('input[type="file"]'),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: function(data) {
				$.toast({
					icon: 'info',
					text: data.message
				});
			},
			error: main.form_ajax_error($f),
			complete: function() {
				main.form_ajax_end($f)();
				job_list();
			}
		});
		return false;
	}

	function job_abort() {
		$('#job_abort').prop('disabled', true);

		var $f = $('#job_form');
		var jid = $('#job_list').find('li.P, li.R').attr('id').replace('job_', '');

		$.ajax({
			url: './abort',
			type: 'POST',
			data: {
				_token_: main.token,
				jid: jid
			},
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: function(data) {
				$.toast({
					icon: 'info',
					text: data.message
				});
			},
			error: main.form_ajax_error($f),
			complete: function() {
				main.form_ajax_end($f)();
				job_list();
			}
		});
		return false;
	}

	function job_status(jid) {
		var $job = $('#job_' + jid);

		if ($job.data('timer')) {
			clearTimeout($job.data('timer'));
			$job.data('timer', null);
		}

		if ($job.find('a.active').length == 0) {
			return;
		}

		var timeout = 0, param = { jid: jid, asc: false, limit: 1000 };
		var $logs = $('#job_logs_' + jid), $tb = $logs.find('tbody'), $tr = $tb.children('tr:first-child');

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

				$job.data('job', job).removeClass('A C P R').addClass(job.status);
				$job.find('i').attr('class', job_status_icon(job.status));

				job_btn_refresh();

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
		var jid = $tb.closest('table').attr('id').replace('job_logs_', '');
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
		var tr = document.createElement("tr"),
			td1 = document.createElement("td"),
			td2 = document.createElement("td"),
			td3 = document.createElement("td");
		tr.className = lg.level;
		tr.setAttribute('data-lid', lg.id);
		td1.textContent = main.format_time(lg.time);
		td2.textContent = '[' + lg.level + ']';
		td3.textContent = lg.message;
		tr.append(td1, td2, td3);
		return tr;
	}

	function job_btn_refresh() {
		var $jobs = $('#job_list'), $ul = $jobs.children('ul'), jpr = $ul.find('li.P, li.R').length > 0;
		$('#job_start').prop('disabled', jpr).find('i')[jpr ? 'addClass' : 'removeClass']('fa-spinner fa-spin');
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
		var jid = $(e.target).attr('href').replace('#job_info_', '');
		job_status(jid);
		return false;
	}

	function build_job_list(data) {
		if (!data || data.length == 0) {
			return;
		}

		var $jobs = $('#job_list'), $ul = $jobs.children('ul');
		if ($ul.length == 0) {
			$ul = $('<ul class="nav nav-pills">');
			$jobs.append($ul);
			$ul.on('click', 'a', job_tab_show);
			$ul.on('shown.bs.tab', 'a', job_tab_shown);
		}

		var $tabs = $jobs.children('div');
		if ($tabs.length == 0) {
			$tabs = $('<div class="tab-content">');
			$jobs.append($tabs);
		}

		for (var i = data.length - 1; i >= 0; i--) {
			var job = data[i], $li = $ul.children('#job_' + job.id);
			if ($li.length) {
				continue;
			}

			var $a = $('<a>', { href: '#job_info_' + job.id, 'class': 'nav-link' });
			$a.append($('<i>', { 'class': job_status_icon(job.status) }));
			$a.append($('<span>').text(main.format_time(job.created_at)));
			
			$li = $('<li>', { id: 'job_' + job.id, 'class': 'nav-item ' + job.status }).append($a);
			$ul.prepend($li.data('job', job));

			var $tab = $('<div>', { id: 'job_info_' + job.id, 'class': "tab-pane fade" });
			if (job.file || job.param) {
				$tab.append(build_job_param(job));
			}
			$tab.append($('<table>', { id: 'job_logs_' + job.id, 'class': "table table-striped"}).append($('<tbody>')));

			$tabs.prepend($tab);
		}

		$ul.find('a').first().trigger('click');
	}

	function job_copy_param() {
		var $form = $(this).closest('form');

		$form.find(':input').prop('disabled', false);
		var vs = $form.formValues();
		$form.find(':input').prop('disabled', true);

		$('#job_form').formValues(vs, true);
		return false;
	}

	function build_job_param(job) {
		var $form = $('#job_form'), legend = $form.prev('legend').text();
		if (!legend) {
			return $('<hr/>');
		}

		var $fset = $('<fieldset>', { 'class': "ui-fieldset collapsed" }).append($('<legend>').text(legend));

		$form = $form.clone().attr('id', 'job_param_' + job.id);
		$form.on('submit', function() { return false; });
		$form.find('[type=hidden]').remove();
		$form.find(':input').prop('disabled', true);
		$form.formClear();

		var $f = $form.find('[type=file]').hide();
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
			$form.formValues(params);

			var $a = $('<a>', { 'class': 'btn btn-secondary ps', href: '#' });
			$a.append($('<i class="fa fa-arrow-up">'));
			$a.append($('<span>').text('Copy'));
			$a.click(job_copy_param);
	
			$form.append($a);
		}

		$form.find('select[data-spy="niceSelect"]').niceSelect();

		$fset.append($form).fieldset().on('expanded.fieldset', function() {
			$(this).find('textarea').autosize();
		});
		return $fset;
	}

	function job_list() {
		var $f = $('#job_form');

		$.ajax({
			url: './list',
			type: 'GET',
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: function(data) {
				setTimeout(function() {
					build_job_list(data);
					job_btn_refresh();
				}, 10);
			},
			error: main.form_ajax_error($f),
			complete: main.form_ajax_end($f)
		});

		return false;
	}

	$('#job_form').submit(function() { return false; });
	$('#job_start').click(job_start);
	$('#job_abort').click(job_abort);

	job_list();
});
