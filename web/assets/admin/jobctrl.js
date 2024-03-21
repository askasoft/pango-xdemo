$(function() {
	function job_start() {
		$('#job_start').prop('disabled', true);

		xdemo.loadmask();
		$.ajaf({
			url: './start',
			type: 'POST',
			data: $('#job_form').serializeArray(),
			file: $('#job_form').find('input[type="file"]'),
			dataType: 'json',
			success: function(data, ts, xhr) {
				$.toast({
					icon: 'info',
					text: data.message
				});
			},
			error: xdemo.ajax_error,
			complete: function() {
				xdemo.unloadmask();
				job_list();
			}
		});
		return false;
	}

	function job_abort() {
		var jid = $('#job_list').find('li.P, li.R').attr('id').replace('job_', '');

		$('#job_abort').prop('disabled', true);

		xdemo.loadmask();
		$.ajax({
			url: './abort',
			type: 'POST',
			data: {
				_token_: xdemo.token,
				jid: jid
			},
			dataType: 'json',
			success: function(data, ts, xhr) {
				$.toast({
					icon: 'info',
					text: data.message
				});
			},
			error: xdemo.ajax_error,
			complete: function() {
				xdemo.unloadmask();
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

		var $logs = $('#job_logs_' + jid),
			$tb = $logs.find('tbody'),
			timeout = 0,
			param = { jid: jid, skip: $tb.children().length, limit: 1000 };

		$.ajax({
			url: './status',
			data: param,
			type: 'GET',
			dataType: 'json',
			success: function(data, ts, xhr) {
				var job = data.job, logs = data.logs || [];

				$job.data('job', job).removeClass('A C P R').addClass(job.status);
				$job.find('i').attr('class', job_status_icon(job.status));

				var jpr = (job.status == 'P' || job.status == 'R');
				var next = jpr || logs.length >= param.limit;

				var updated_at = new Date(job.updated_at);
				if (next || (new Date().getTime() - updated_at.getTime() < 10000)) {
					timeout = 1000;
				}

				if (logs.length > 0) {
					var tb = $tb.get(0), c = document.createDocumentFragment();;
					for (var i = logs.length - 1; i >= 0; i--) {
						var lg = logs[i],
							tr = document.createElement("tr"),
							td1 = document.createElement("td"),
							td2 = document.createElement("td"),
							td3 = document.createElement("td");
						tr.className = lg.level;
						td1.textContent = xdemo.format_time(lg.when);
						td2.textContent = '[' + lg.level + ']';
						td3.textContent = lg.message;
						tr.append(td1, td2, td3);
						c.append(tr);
						if (i >= logs.length - 50) {
							tr.className = lg.level + " hidden";
						}
					}
					tb.prepend(c);

					if (!$logs.data('timer')) {
						$logs.data('timer', setTimeout(function() { show_log(jid); }, 10));
					}
				}

				job_btn_refresh();

				if (timeout) {
					$job.data('timer', setTimeout(function() { job_status(jid); }, timeout));
				}
			},
			error: xdemo.ajax_error
		});
	}

	function show_log(jid) {
		var $logs = $('#job_logs_' + jid), $trs = $logs.find('tr.hidden');

		$logs.data('timer', null);
		if ($trs.length == 0) {
			return;
		}

		$trs.last().removeClass('hidden');
		if ($trs.length > 0) {
			$logs.data('timer', setTimeout(function() { show_log(jid); }, 20));
		}
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
			$a.append($('<span>').text(xdemo.format_time(job.created_at)));
			
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
		var $form = $('#job_form'), $legend = $form.prev('legend');
		var $fset = $('<fieldset>', { 'class': "ui-fieldset collapsed" }).append($('<legend>').text($legend.text()));

		$form = $form.clone().attr('id', 'job_param_' + job.id);
		$form.on('submit', function() { return false; });
		$form.find('[type=hidden]').remove();
		$form.find(':input').prop('disabled', true);
		$form.formClear();

		var $f = $form.find('[type=file]').hide();
		if (job.file) {
			$('<a>', { 'class': 'btn btn-secondary', href: xdemo.base + '/files' + job.file })
				.append($('<i>', { 'class': 'fas fa-download' }))
				.insertAfter($f);
		}

		if (job.param) {
			var params = JSON.parse(job.param);
			for (var k in params) {
				var v = params[k];
				if (typeof(v) == 'string') {
					var d = new Date(v);
					if (d.getTime() > 0 ) {
						params[k] = xdemo.format_date(d);
					}
				}
			}
			$form.formValues(params);

			var $a = $('<a>', { 'class': 'btn btn-outline-secondary ps', href: '#' });
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
		$.ajax({
			url: './list',
			type: 'GET',
			dataType: 'json',
			success: function(data, ts, xhr) {
				build_job_list(data);
				job_btn_refresh();
			},
			error: xdemo.ajax_error
		});
		return false;
	}

	$('#job_form').submit(function() { return false; });
	$('#job_start').click(job_start);
	$('#job_abort').click(job_abort);

	job_list();
});
