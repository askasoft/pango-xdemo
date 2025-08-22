(function($) {
	var langs = {
		en: {
			units: {
				"": "Select...",
				d: "Daily", 
				w: "Weekly", 
				m: "Monthly"
			},
			dows: [ "MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN" ],
			doms: [ "Day 1", "Day 2", "Day 3", "Day 4", "Day 5", "Day 6", "Day 7", "Day 8", "Day 9", "Day 10", "Day 11", "Day 12", "Day 13", "Day 14", "Day 15", "Day 16", "Day 17", "Day 18", "Day 19", "Day 20", "Day 21", "Day 22", "Day 23", "Day 24", "Day 25", "Day 26", "Day 27", "Day 28", "Day 29", "Day 30", "Day 31" ],
			hours: [ "12 AM", "1 AM", "2 AM", "3 AM", "4 AM", "5 AM", "6 AM", "7 AM", "8 AM", "9 AM", "10 AM", "11 AM", "12 PM", "1 PM", "2 PM", "3 PM", "4 PM", "5 PM", "6 PM", "7 PM", "8 PM", "9 PM", "10 PM", "11 PM" ]
		},
		ja: {
			units: {
				"": "選択...",
				d: "毎日",
				w: "毎週",
				m: "毎月"
			},
			dows: [ "月曜日", "火曜日", "水曜日", "木曜日", "金曜日", "土曜日", "日曜日" ],
			doms: [ "1日", "2日", "3日", "4日", "5日", "6日", "7日", "8日", "9日", "10日", "11日", "12日", "13日", "14日", "15日", "16日", "17日", "18日", "19日", "20日", "21日", "22日", "23日", "24日", "25日", "26日", "27日", "28日", "29日", "30日", "31日" ],
			hours: [ "0時", "1時", "2時", "3時", "4時", "5時", "6時", "7時", "8時", "9時", "10時", "11時", "12時", "13時", "14時", "15時", "16時", "17時", "18時", "19時", "20時", "21時", "22時", "23時" ]
		},
		zh: {
			units: {
				"": "选择...",
				d: "每天",
				w: "每周",
				m: "每月"
			},
			dows: [ "周一", "周二", "周三", "周四", "周五", "周六", "周日" ],
			doms: [ "1日", "2日", "3日", "4日", "5日", "6日", "7日", "8日", "9日", "10日", "11日", "12日", "13日", "14日", "15日", "16日", "17日", "18日", "19日", "20日", "21日", "22日", "23日", "24日", "25日", "26日", "27日", "28日", "29日", "30日", "31日" ],
			hours: [ "0时", "1时", "2时", "3时", "4时", "5时", "6时", "7时", "8时", "9时", "10时", "11时", "12时", "13时", "14时", "15时", "16时", "17时", "18时", "19时", "20时", "21时", "22时", "23时" ]
		}
	};

	//----------------------------------------------------
	function schedule_field_update_value($i) {
		if ($i.length == 0 || $i.prop('disabled')) {
			return;
		}

		var $a = $i.parent().find('.schedule');
		var unit = $a.find('.unit').val();
		var dow = $a.find('.dow').val();
		var dom = $a.find('.dom').val();
		var hour = $a.find('.hour').val();
		var v = '';
		switch (unit) {
		case 'd':
			v = unit + ' 0 ' + hour;
			break;
		case 'w':
			v = unit + ' ' + dow + ' ' + hour;
			break;
		case 'm':
			v = unit + ' ' + dom + ' ' + hour;
			break;
		}
		$i.val(v);
	}

	function schedule_field_onchange() {
		var $s = $(this);
		if ($s.hasClass('unit')) {
			var v = $s.val(), $p = $s.parent();
			$p.find('.dow').toggleClass('hidden', v != 'w');
			$p.find('.dom').toggleClass('hidden', v != 'm');
			$p.find('.hour').toggleClass('hidden', v != 'd' && v != 'w' && v != 'm');
		}

		var $i = $(this).closest('.schedule').parent().find('input[type="hidden"]');
		schedule_field_update_value($i);
		return false;
	}

	function schedule_field_init($i) {
		if ($i.length == 0) {
			return;
		}

		var ln = $('html').attr('lang') || 'en';
		var ss = $i.val().split(' ');
		var unit = ss.length ? ss[0] : '';
		var day = ss.length > 1 ? parseInt(ss[1]) : 0;
		var hour = ss.length > 2 ? parseInt(ss[2]) : 0;

		if (hour < 0 || hour > 23) {
			hour = 0;
		}

		switch (unit) {
		case 'd':
			break;
		case 'w':
			if (day < 1 || day > 7) {
				day = 1;
			}
			break;
		case 'm':
			if (day < 1 || day > 31) {
				day = 1;
			}
			break;
		}

		var d = !!($i.prop('disabled'));

		var $a = $('<div class="schedule"></div>');
		var $u = $('<select class="unit form-select">').prop('disabled', d);
		var units = langs[ln].units;
		for (var i in units) {
			$u.append($('<option>').val(i).text(units[i]));
		}
		$u.val(unit);

		var $dow = $('<select class="dow form-select hidden">').prop('disabled', d);
		var dows = langs[ln].dows;
		for (var i = 1; i <= dows.length; i++) {
			$dow.append($('<option>').val(i).text(dows[i-1]));
		}
		if (unit == 'w') {
			$dow.val(day).removeClass('hidden');
		}

		var $dom = $('<select class="dom form-select hidden">').prop('disabled', d);
		var doms = langs[ln].doms;
		for (var i = 1; i <= doms.length; i++) {
			$dom.append($('<option>').val(i).text(doms[i-1]));
		}
		if (unit == 'm') {
			$dom.val(day).removeClass('hidden');
		}

		var $hour = $('<select class="hour form-select hidden">').prop('disabled', d);
		var hours = langs[ln].hours;
		for (var i = 0; i < hours.length; i++) {
			$hour.append($('<option>').val(i).text(hours[i]));
		}
		if (unit == 'd' || unit == 'w' || unit == 'm') {
			$hour.val(hour).removeClass('hidden');
		}

		$i.parent().append($a.append($u, $dow, $dom, $hour).on('change', 'select', schedule_field_onchange));
	}

	function schedule_fields_init() {
		schedule_field_init($('[name=schedule_pets_reset]'));
	}

	//----------------------------------------------------
	function toggle_login_method() {
		var $i = $('input[name="secure_login_method"]:checked'), v = $i.val();
		var ldap = v == 'L', saml = v == 'S', pass = !ldap && !saml;
		$('[name="secure_login_mfa"]').prop('disabled', !pass).closest('.row')[pass ? 'slideUp' : 'slideDown']();
		$('[name^="secure_ldap_"]').prop('disabled', !ldap).closest('.row')[ldap ? 'slideDown' : 'slideUp']();
		$('[name^="secure_saml_"]').prop('disabled', !saml).closest('.row')[saml ? 'slideDown' : 'slideUp']();
		$i.closest('.cfgform').find('textarea').autosize();
	}

	//----------------------------------------------------
	function configs_import() {
		var $p = $('#configs_import_popup').popup('update', { keyboard: false });

		$.ajaf({
			url: './import',
			method: 'POST',
			data: $p.find('form').serializeArray(),
			file: $p.find('input[type="file"]'),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success,
				});

				setTimeout(function() {
					location.reload();
				}, 3000);
			},
			error: main.ajax_error,
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});

		return false;
	}

	function configs_export() {
		$.ajaf({
			url: './export',
			method: 'POST',
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}

	function configs_save() {
		var $f = $(this);

		$.ajax({
			url: './save',
			method: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: main.ajax_success,
			error: main.form_ajax_error($f),
			complete: main.form_ajax_end($f)
		});

		return false;
	}

	function configs_tab_show(id) {
		var $t = $('a[href="#' + id + '"]');
		if ($t.length) {
			new bootstrap.Tab($t.get(0)).show();
		}
	}

	function configs_init() {
		schedule_fields_init();

		$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function() {
			var t = $(this).attr('href');
			$(t).find('textarea').autosize();
			history.replaceState(null, null, location.href.split('#')[0] + t);
		});

		$('.cfgform').find('textarea').autosize();
		$('.cfgform').on('submit', configs_save);

		$('#configs_export').on('click', configs_export);
		$('#configs_import_popup').on('click', 'button[type=submit]', configs_import);

		var cg = location.hash.substrAfter('#'), cc = cg;
		if (cc.startsWith('cg_')) {
			cc = $('#' + cc).parent().closest('.tab-pane').attr('id');
		}
		if (cc.startsWith('cc_')) {
			configs_tab_show(cc);
		}
		if (cg.startsWith('cg_')) {
			configs_tab_show(cg);
		}

		$('.cfgform input[name="secure_login_method"]').on('change', toggle_login_method).trigger('change');
	}

	$(window).on('load', configs_init);
})(jQuery);
