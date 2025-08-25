(function($) {
	var langs = {
		en: {
			units: {
				d: "Daily", 
				w: "Weekly", 
				m: "Monthly"
			},
			dows: [ "SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT", "" ],
			doms: [ "", "Day 1", "Day 2", "Day 3", "Day 4", "Day 5", "Day 6", "Day 7", "Day 8", "Day 9", "Day 10", "Day 11", "Day 12", "Day 13", "Day 14", "Day 15", "Day 16", "Day 17", "Day 18", "Day 19", "Day 20", "Day 21", "Day 22", "Day 23", "Day 24", "Day 25", "Day 26", "Day 27", "Day 28", "Day 29", "Day 30", "Day 31", "Last" ],
			hours: [ "12 AM", "1 AM", "2 AM", "3 AM", "4 AM", "5 AM", "6 AM", "7 AM", "8 AM", "9 AM", "10 AM", "11 AM", "12 PM", "1 PM", "2 PM", "3 PM", "4 PM", "5 PM", "6 PM", "7 PM", "8 PM", "9 PM", "10 PM", "11 PM" ]
		},
		ja: {
			units: {
				d: "毎日",
				w: "毎週",
				m: "毎月"
			},
			dows: [ "", "月曜日", "火曜日", "水曜日", "木曜日", "金曜日", "土曜日", "日曜日" ],
			doms: [ "", "1日", "2日", "3日", "4日", "5日", "6日", "7日", "8日", "9日", "10日", "11日", "12日", "13日", "14日", "15日", "16日", "17日", "18日", "19日", "20日", "21日", "22日", "23日", "24日", "25日", "26日", "27日", "28日", "29日", "30日", "31日", "月末" ],
			hours: [ "0時", "1時", "2時", "3時", "4時", "5時", "6時", "7時", "8時", "9時", "10時", "11時", "12時", "13時", "14時", "15時", "16時", "17時", "18時", "19時", "20時", "21時", "22時", "23時" ]
		},
		zh: {
			units: {
				d: "每天",
				w: "每周",
				m: "每月"
			},
			dows: [ "", "周一", "周二", "周三", "周四", "周五", "周六", "周日" ],
			doms: [ "", "1日", "2日", "3日", "4日", "5日", "6日", "7日", "8日", "9日", "10日", "11日", "12日", "13日", "14日", "15日", "16日", "17日", "18日", "19日", "20日", "21日", "22日", "23日", "24日", "25日", "26日", "27日", "28日", "29日", "30日", "31日", "月末" ],
			hours: [ "0时", "1时", "2时", "3时", "4时", "5时", "6时", "7时", "8时", "9时", "10时", "11时", "12时", "13时", "14时", "15时", "16时", "17时", "18时", "19时", "20时", "21时", "22时", "23时" ]
		}
	};

	//----------------------------------------------------
	function schedule_field_update_value($s) {
		var v = $s.find('.units input:checked').val() || '';
		var dows = $s.find('.dows input:checked').map(function() { return $(this).val(); }).get().join(',');
		var doms = $s.find('.doms input:checked').map(function() { return $(this).val(); }).get().join(',');
		var hours = $s.find('.hours input:checked').map(function() { return $(this).val(); }).get().join(',');

		var unit = v.substring(0, 1);
		switch (unit) {
		case 'd':
			v = unit + ' * ' + hours;
			break;
		case 'w':
			v = unit + ' ' + dows + ' ' + hours;
			break;
		case 'm':
			v = unit + ' ' + doms + ' ' + hours;
			break;
		}

		if (unit) {
			$s.find('.units .' + unit).val(v);
		}
	}

	function schedule_field_onchange() {
		var $i = $(this), $u = $i.closest('.units'), $s = $i.closest('.schedule');
		if ($u.length) {
			var u = ($u.find('input:checked').val() || '').substring(0, 1);
			$s.find('.dows').toggleClass('hidden', u != 'w');
			$s.find('.doms').toggleClass('hidden', u != 'm');
			$s.find('.hours').toggleClass('hidden', u != 'd' && u != 'w' && u != 'm');
		}

		schedule_field_update_value($s);
		return false;
	}

	function schedule_field_init($i) {
		if ($i.length == 0) {
			return;
		}

		var nm = $i.attr('name');
		var ln = langs[$('html').attr('lang')] || langs.en;
		var ss = $i.val().split(' ');
		var unit = ss.length ? ss[0] : '';
		var days = ss.length > 1 ? ss[1].split(',') : '';
		var hours = ss.length > 2 ? ss[2].split(',') : '';

		var d = !!($i.prop('disabled'));

		var $s = $('<div class="schedule"></div>');
		var $u = $('<div class="units ui-radios">').append($('<label>').append(
			$('<input type="radio" checked>').attr('name', nm).prop('disabled', d),
			$('<span>').text('ー')
		));
		for (var i in ln.units) {
			$u.append($('<label>').append(
				$('<input type="radio">').attr('name', nm).addClass(i).val(i).prop('disabled', d),
				$('<span>').text(ln.units[i])
			));
		}
		$u.find('[value="' + unit +'"]').prop('checked', true).val($i.val());

		var $dows = $('<div class="dows ui-checks hidden">');
		$.each(ln.dows, function(i, dow) {
			if (dow) {
				$dows.append($('<label>').append(
					$('<input type="checkbox">').val(i).prop('checked', days.indexOf(i+'') >= 0).prop('disabled', d),
					$('<span>').text(dow)
				));
			}
		});
		if (unit == 'w') {
			$dows.removeClass('hidden');
		}

		var $doms = $('<div class="doms ui-checks hidden">');
		$.each(ln.doms, function(i, dom) {
			if (dom) {
				$doms.append($('<label>').append(
					$('<input type="checkbox">').val(i).prop('checked', days.indexOf(i+'') >= 0).prop('disabled', d),
					$('<span>').text(dom)
				));
			}
		});
		if (unit == 'm') {
			$doms.removeClass('hidden');
		}

		var $hours = $('<div class="hours ui-checks hidden">');
		$.each(ln.hours, function(i, hour) {
			$hours.append($('<label>').append(
				$('<input type="checkbox">').val(i).prop('checked', hours.indexOf(i+'') >= 0).prop('disabled', d),
				$('<span>').text(hour)
			));
		});
		if (unit == 'd' || unit == 'w' || unit == 'm') {
			$hours.removeClass('hidden');
		}

		$s.append($u, $dows, $doms, $hours).on('change', 'input', schedule_field_onchange);
		$s.insertAfter($i);
		$i.remove();
	}

	function schedule_fields_init() {
		schedule_field_init($('[name=schedule_pets_reset]'));
	}

	//----------------------------------------------------
	function secure_login_method_init() {
		$('.cfgform input[name="secure_login_method"]').on('change', function() {
			var $i = $('input[name="secure_login_method"]:checked'), v = $i.val();
			var ldap = v == 'L', saml = v == 'S';
			$('[name="secure_login_mfa"]').prop('disabled', saml).closest('.row')[saml ? 'slideUp' : 'slideDown']();
			$('[name^="secure_ldap_"]').prop('disabled', !ldap).closest('.row')[ldap ? 'slideDown' : 'slideUp']();
			$('[name^="secure_saml_"]').prop('disabled', !saml).closest('.row')[saml ? 'slideDown' : 'slideUp']();
			$i.closest('.cfgform').find('textarea').autosize();
		}).trigger('change');
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
		secure_login_method_init();
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
			cc = $('#' + cc).parent().closest('.tab-pane').attr('id') || '';
		}
		if (cc.startsWith('cc_')) {
			configs_tab_show(cc);
		}
		if (cg.startsWith('cg_')) {
			configs_tab_show(cg);
		}
	}

	$(window).on('load', configs_init);
})(jQuery);
