var main = {
	base: '',
	cookie: { expires: 180 },

	meta_props: function() {
		var m = {}, p = 'xdemo:';
		$('meta').each(function() {
			var $t = $(this), a = $t.attr('property');
			if (a && a.substring(0, p.length) == p) {
				m[a.substring(p.length)] = $t.attr('content');
			}
		});
		return m;
	},

	fmt_date: new DateFormat("yyyy-MM-dd"),
	fmt_time: new DateFormat("yyyy-MM-dd HH:mm:ss"),
	format_date: function(d) {
		if (typeof(d) == 'string') {
			d = new Date(d);
		}
		return main.fmt_date.format(d);
	},
	format_time: function(d) {
		if (typeof(d) == 'string') {
			d = new Date(d);
		}
		return main.fmt_time.format(d);
	},

	isURL: function(s) {
		return /^https?:\/\/[\w~!@#\$%&\*\(\)_\-\+=\[\]\|:;,\.\?\/']+$/i.test(s + '');
	},

	safe_parse_json: function(s, d) {
		try {
			return $.parseJSON(s);
		} catch (e) {
			console.error(e);
			return d || s;
		}
	},

	// session storage
	ssload: function(k) {
		try {
			var s = sessionStorage[k];
			if (s) {
				return JSON.parse(s);
			}
		} catch (e) {
			console.error(e);
		}
		return {};
	},
	sssave: function(k, o) {
		try {
			sessionStorage[k] = JSON.stringify(o);
		} catch (e) {
			console.error(e);
		}
	},

	// local storage
	lsload: function(k) {
		try {
			var s = localStorage[k];
			if (s) {
				return JSON.parse(s);
			}
		} catch (e) {
			console.error(e);
		}
		return {};
	},
	lssave: function(k, o) {
		try {
			localStorage[k] = JSON.stringify(o);
		} catch (e) {
			console.error(e);
		}
	},

	// replace location search
	location_replace_search: function(vs) {
		var p = $.extend({}, vs);

		delete p.p; // page
		delete p.l; // limit
		delete p.c; // sort column
		delete p.d; // sort direction

		var q = $.param(p, true);
		history.replaceState(null, null, location.href.split('?')[0] + (q ? '?' + q : ''));
	},

	// load mask
	loadmask: function() {
		$('body').loadmask();
	},
	unloadmask: function() {
		$('body').unloadmask();
	},

	// ajax setup token header
	ajax_setup: function() {
		$.ajaxPrefilter(function(options, original, xhr) {
			xhr.setRequestHeader('X-CSRF-TOKEN', main.token);
		});
	},

	// ajaf error handler
	ajaf_error: function(data) {
		data = main.safe_parse_json(data);
		if (data && data.error) {
			$.toast({
				icon: 'error',
				text: data.error
			});
			return true;
		}
		return false;
	},

	// ajax error handler
	ajax_error: function(xhr, status, err, $f) {
		if (xhr.readyState == 0) { // window unload
			console.log('ajax canceled.');
			return;
		}

		var afterHidden;
		if (xhr.status == 401 || xhr.status == 403) { // unauthorized, forbidden
			afterHidden = function() {
				location.href = main.base + '/login/?origin=' + encodeURIComponent(location.pathname + location.search + location.hash);
			};
		}

		err = (xhr.status ? xhr.status + ' ' : '') + (err || status || 'error');
		if (xhr.responseJSON) {
			err = xhr.responseJSON.error || JSON.stringify(xhr.responseJSON, null, 2) || err;
		}

		if (!$.isArray(err)) {
			err = [ err ];
		}

		var msgs = [];
		$.each(err, function(i, e) {
			if (e.param) {
				main.form_add_invalid($f, e);
			}
			msgs.push((e.label ? e.label + ': ' : '') + (e.message || e + ""));
		});

		if (msgs.length == 1) {
			msgs = msgs[0];
		}

		$.toast({
			icon: 'error',
			text: msgs,
			afterHidden: afterHidden
		});
	},
	ajax_success: function(data) {
		if (data.success) {
			$.toast({
				icon: 'success',
				text: data.success
			});
		}
		if (data.warning) {
			$.toast({
				icon: 'warning',
				text: data.warning
			});
		}
	},
	popup_ajax_fail: function(paf) {
		return function(xhr, status, err) {
			main.ajax_error(xhr, status, err);
			paf.call(this, xhr, status, err);
		}
	},

	form_input_values: function($f) {
		var vs = $f.formValues();
		for (var k in vs) {
			if ($f.find('[name="'+ k +'"]').prop('type') == 'checkbox') {
				// do not delete empty check value
				continue;
			}

			if (!vs[k]) {
				delete vs[k];
			}
		}
		return vs;
	},
	form_has_inputs: function($f) {
		var f = function() {
			var $i = $(this);
			if ($i.hasClass('ignore') || $i.attr('type') == 'hidden') {
				return false;
			}
			if ($i.is(':checkbox')) {
				return $i.is(':checked');
			}
			var v = $i.val();
			return v && v.length > 0;
		};
		return $f.find('input, select').filter(f).length;
	},
	form_clear_invalid: function($f) {
		$f.find('.is-invalid').removeClass('is-invalid').end().find('.verr').remove();
	},
	form_add_invalid: function($f, err) {
		if ($f && $f.length) {
			var $i = $f.find('[name="' + err.param + '"]');
			$i.addClass('is-invalid');
			$i.eq(0).closest('div').append($('<div class="verr">').text(err.message));
		}
	},
	form_ajax_error: function($f) {
		return function(xhr, status, err) {
			main.ajax_error(xhr, status, err, $f);
		};
	},
	form_ajax_start: function($f) {
		return function() {
			main.form_clear_invalid($f);
			$f.loadmask();
		};
	},
	form_ajax_end: function($f) {
		return function() {
			$f.unloadmask();
		};
	},

	// show alert
	show_alert: function($c, type, msgs) {
		var icons = {
			'danger': 'fa-circle-exclamation',
			'warning': 'fa-triangle-exclamation',
			'success': 'fa-check',
			'primary': 'fa-circle-info'
		};

		var $a = $('<div class="alert alert-primary alert-dismissible"></div>').appendTo($c);
		$('<button type="button" class="btn-close" data-bs-dismiss="alert"></button>').appendTo($a);

		var $ul = $('<ul class="fa-ul"></ul>').appendTo($a);
		var add = function(i, msg) {
			var $li = $('<li><span class="fa-li"><i class="fas fa-fw ' + icons[type] + '"></i></span></li>');
			var $sp = $('<span></span>').text(msg);
			$li.append($sp).appendTo($ul);
		};

		if ($.isArray(msgs)) {
			$.each(msgs, add);
		} else {
			add(0, msgs);
		}
	},

	ui_init: function($d) {
		$d.find('[data-spy="fieldset"]').fieldset();
		$d.find('[data-spy="niceSelect"]').niceSelect();
		$d.find('[data-spy="uploader"]').uploader();
		$d.find('[textclear]').textclear();
		$d.find('textarea[autosize]').autosize();
		$d.find('textarea[enterfire]').enterfire();
		$d.find('[linkify]').linkify();
		main.summernote($d.find('textarea[summernote]'));
	},

	// list
	list_init: function(name, sskey) {
		var $l = $('#' + name + '_list'), $f = $('#' + name + '_listform'), tb = '#' + name + '_table';

		if (!location.search) {
			$f.formValues(main.ssload(sskey), true);
		}
		if (main.form_has_inputs($f)) {
			$('#' + name + '_listfset').fieldset('expand', 'show');
		}

		$l.on('goto.pager', '.ui-pager', function(evt, pno) {
			$f.find('input[name="p"]').val(pno).end().submit();
		});
		$l.on('change', '.ui-pager select', function() {
			$f.find('input[name="l"]').val($(this).val()).end().submit();
		});
		$l.on('sort.sortable', tb, function(evt, col, dir) {
			$f.find('input[name="c"]').val(col);
			$f.find('input[name="d"]').val(dir);
			$f.submit();
		});
	},
	list_build: function($l, data) {
		$l.html(data);

		$l.find('[checkall]').checkall();
		$l.find('[data-spy="pager"]').pager();
		$l.find('[data-spy="sortable"]').sortable();
	},
	list_builder: function($l, callback) {
		return function(data) {
			main.list_build($l, data);
			if (callback) {
				setTimeout(callback, 100);
			}
		};
	},

	// detail popup
	detail_popup_shown: function($d) {
		$d.find('.ui-popup-body').prop('scrollTop', 0);
		main.ui_init($d);
		$(window).trigger('resize');
	},
	detail_popup_shown2: function($d) {
		$d.find('.ui-popup-body > *').prop('scrollTop', 0);
		main.detail_popup_shown($d);
	},
	detail_popup_prevnext: function($d, $l, idpx) {
		var $p = $d.find('.prev'), $n = $d.find('.next');

		var id = $d.find('input[name=id]').val(), $tr = $(idpx + id), $pg = $l.find('.ui-pager > .pagination');
		var prev = $tr.prev('tr').length || $pg.find('.page-item.prev.disabled').length == 0;
		var next = $tr.next('tr').length || $pg.find('.page-item.next.disabled').length == 0;

		$p.toggle(id != '0' && prev);
		$n.toggle(id != '0' && next);
	},
	detail_popup_keydown: function(evt) {
		if (evt.altKey) {
			switch (evt.key) {
			case 'ArrowLeft':
				evt.preventDefault();
				var $p = $(this).children('.prev');
				if (!$p.is(':hidden')) {
					$p.trigger('click');
				}
				break;
			case 'ArrowRight':
				evt.preventDefault();
				var $n = $(this).children('.next');
				if (!$n.is(':hidden')) {
					$n.trigger('click');
				}
				break;
			}
		}
	},

	// bulk edit
	bulkedit_editsel_popup: function(name) {
		var ids = main.get_table_checked_ids($('#' + name + '_table'));
		$('#' + name + '_bulkedit_popup')
			.find('.editsel').show().end()
			.find('.editall').hide().end()
			.find('input[name=id]').val(ids.join(',')).end()
			.popup('show');
		return false;
	},
	bulkedit_editall_popup: function(name) {
		$('#' + name + '_bulkedit_popup')
			.find('.editsel').hide().end()
			.find('.editall').show().end()
			.find('input[name=id]').val('*').end()
			.popup('show');
		return false;
	},
	bulkedit_label_click: function() {
		var $t = $(this), $i = $t.parent().next().find(':input');
		$i.prop('disabled', !$t.prop('checked'));
		if ($i.data('spy') == 'niceSelect') {
			$i.niceSelect('update');
		}
		// NOTE: do not 'return false', otherwise 'update' button can not be enabled.
	},

	// table
	get_table_trs: function(px, ids) {
		var trs = [];
		$.each(ids, function(i, v) {
			trs.push(px + v);
		});
		return $(trs.join(','));
	},
	get_table_checked_ids: function($tb) {
		var ids = [];
		$tb.find('td.check > input:checked').each(function() {
			ids.push($(this).val());
		});
		return ids;
	},
	set_table_tr_values: function($tr, vs) {
		for (var k in vs) {
			var $td = $tr.find('td.' + k);
			if ($td.length == 0 || $td.hasClass('ro')) {
				continue;
			}

			var $c = $td.children('a, pre, button'), v = vs[k];
			if (typeof(v) == 'undefined') {
				v = '';
			}
			if (v && k.endsWith("_at")) {
				v = main[$td.hasClass('date') ? 'format_date' : 'format_time'](v);
			}
			if ($c.length) {
				var $s = $c.children('s');
				($s.length ? $s : $c).text(v);
			} else {
				$td.text(v);
			}
		}
	},

	// blink
	blink_start: function($e) {
		$e.addClass('ui-blink-1s');
	},
	blink_stop: function($e) {
		$e.removeClass('ui-blink-1s');
	},
	blink: function($e) {
		$e.addClass('ui-blink-1s2');
		setTimeout(function() { $e.removeClass('ui-blink-1s2'); }, 2000);
	},
	summernoteLangs: {
		"ar-AR": "ar-AR",
		"az-AZ": "az-AZ",
		"bg-BG": "bg-BG",
		"bn-BD": "bn-BD",
		"ca-ES": "ca-ES",
		"cs-CZ": "cs-CZ",
		"da-DK": "da-DK",
		"de-CH": "de-CH",
		"de-DE": "de-DE",
		"el-GR": "el-GR",
		"en-US": "en-US",
		"es-ES": "es-ES",
		"es-EU": "es-EU",
		"fa-IR": "fa-IR",
		"fi-FI": "fi-FI",
		"fr-FR": "fr-FR",
		"gl-ES": "gl-ES",
		"he-IL": "he-IL",
		"hr-HR": "hr-HR",
		"hu-HU": "hu-HU",
		"id-ID": "id-ID",
		"it-IT": "it-IT",
		"ja-JP": "ja-JP",
		"ko-KR": "ko-KR",
		"lt-LT": "lt-LT",
		"lt-LV": "lt-LV",
		"mn-MN": "mn-MN",
		"nb-NO": "nb-NO",
		"nl-NL": "nl-NL",
		"pl-PL": "pl-PL",
		"pt-BR": "pt-BR",
		"pt-PT": "pt-PT",
		"ro-RO": "ro-RO",
		"ru-RU": "ru-RU",
		"sk-SK": "sk-SK",
		"sl-SI": "sl-SI",
		"sr-RS-Latin": "sr-RS-Latin",
		"sr-RS": "sr-RS",
		"sv-SE": "sv-SE",
		"ta-IN": "ta-IN",
		"th-TH": "th-TH",
		"tr-TR": "tr-TR",
		"uk-UA": "uk-UA",
		"uz-UZ": "uz-UZ",
		"vi-VN": "vi-VN",
		"zh-CN": "zh-CN",
		"zh-TW": "zh-TW",
		"ar": "ar-AR",
		"az": "az-AZ",
		"bg": "bg-BG",
		"bn": "bn-BD",
		"ca": "ca-ES",
		"cs": "cs-CZ",
		"da": "da-DK",
		"de": "de-DE",
		"el": "el-GR",
		"en": "en-US",
		"es": "es-ES",
		"es": "es-EU",
		"fa": "fa-IR",
		"fi": "fi-FI",
		"fr": "fr-FR",
		"gl": "gl-ES",
		"he": "he-IL",
		"hr": "hr-HR",
		"hu": "hu-HU",
		"id": "id-ID",
		"it": "it-IT",
		"ja": "ja-JP",
		"ko": "ko-KR",
		"lt": "lt-LT",
		"mn": "mn-MN",
		"nb": "nb-NO",
		"nl": "nl-NL",
		"pl": "pl-PL",
		"pt": "pt-PT",
		"ro": "ro-RO",
		"ru": "ru-RU",
		"sk": "sk-SK",
		"sl": "sl-SI",
		"sr": "sr-RS",
		"sv": "sv-SE",
		"ta": "ta-IN",
		"th": "th-TH",
		"tr": "tr-TR",
		"uk": "uk-UA",
		"uz": "uz-UZ",
		"vi": "vi-VN",
		"zh": "zh-CN",
		"zh-HK": "zh-CN"
	},
	summernote: function($el) {
		if ($el.length) {
			var sno = {
				fontSizes: ['8', '9', '10', '11', '12', '14', '16', '18', '20', '24', '36'],
				toolbar: [
					[ 'style', [ 'style', 'fontname', 'fontsize', 'color' ] ],
					[ 'text', [ 'bold', 'italic', 'underline', 'strikethrough', 'superscript', 'subscript' ] ],
					[ 'para', [ /*'height', */'paragraph', 'ol', 'ul' ] ],
					[ 'insert', [ 'hr', 'table', 'link', 'picture', 'video' ] ],
					[ 'edit', [ 'undo', 'redo', 'clear' ] ],
					[ 'misc', [ 'codeview', 'fullscreen', 'help' ] ]
				]
			};

			var hln = $('html').attr('lang'), sln = main.summernoteLangs[hln];
			if (hln || sln) {
				sno.lang = (sln ? sln : hln);
			}
			$el.removeAttr('summernote').summernote(sno);

			// summernote bs5 bug fix
			// $('.note-toolbar').find('[data-toggle]').each(function() {
			// 	$(this).attr('data-bs-toggle', $(this).attr('data-toggle')).removeAttr('data-toggle');
			// });
		}
	},

	init: function() {
		// summernote
		main.summernote($('textarea[summernote]'));

		// sidenavi
		$('#sidenavi i').each(function() {
			$(this).attr('title', $(this).next('span').text());
		});

		// header locale switch
		$('#header .locale a').on('click', function() {
			var $a = $(this), t = $a.attr('href').substring(1);

			$.cookie('X_LOCALE', t, { path: main.base + '/' });

			location.reload();
			return false;
		});

		// header theme switch
		$('#header .theme a').on('click', function() {
			var $a = $(this), t = $a.attr('href').substring(1);

			localStorage.theme = t;

			$('body').attr('data-bs-theme', t);
			$('#header .theme a').removeClass('active');
			$a.addClass('active');
			return false;
		}).filter('[href="#' + localStorage.theme + '"]').trigger('click');
	}
};

//----------------------------------------------------
(function($) {
	// enable script cache
	$.enableScriptCache();

	// set cookie defaults
	$.extend($.cookie.defaults, main.cookie);

	// set toast defaults
	$.extend($.toast.defaults, {
		position: 'top center'
	});

	// set popup defaults
	$.extend($.popup.defaults, {
		transition: 'zoomIn'
	});

	// get meta properties
	$.extend(main, main.meta_props());

	// ajax setup
	main.ajax_setup();

	// popup setup
	$.popup.defaults.ajaxFail = main.popup_ajax_fail($.popup.defaults.ajaxFail);

	// load init
	$(window).on('load', main.init);
})(jQuery);

