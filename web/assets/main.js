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
				window.location.href = main.base + '/login/';
			};
		}

		err = err || status || 'Server error';
		if (xhr.responseJSON) {
			err = xhr.responseJSON.error || JSON.stringify(xhr.responseJSON, null, 4) || err;
		}

		if ($.isArray(err)) {
			var es = [];
			$.each(err, function(i, e) {
				if (e.param) {
					main.form_add_invalid($f, e);
					es.push(e.message);
				} else {
					es.push(e.message || e + "");
				}
			});
			err = es;
		} else if (err.param && err.message) {
			main.form_add_invalid($f, err);
			err = err.message;
		}

		$.toast({
			icon: 'error',
			text: err,
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

	// list
	list_events: function(name) {
		var $l = $('#' + name + '_list'), $f = $('#' + name + '_listform'), tb = '#' + name + '_table';

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

	// form input (not hidden) values
	form_input_values: function($f) {
		var vs = $f.formValues();

		delete vs._token_;
		for (var k in vs) {
			if (!vs[k]) {
				delete vs[k];
			}
		}
		return vs;
	},
	form_has_inputs: function($f) {
		var f = function() {
			var v = $(this).val();
			return v && v.length > 0;
		};
		return $f.find('input:checked').length
			|| $f.find('select').filter(f).length
			|| $f.find('input[type=text]').filter(f).length;
	},
	form_clear_invalid: function($f) {
		$f.find('.is-invalid').removeClass('is-invalid').end().find('.verr').remove();
	},
	form_add_invalid: function($f, err) {
		if ($f && $f.length) {
			var $i = $f.find('[name="' + err.param + '"]');
			$i.addClass('is-invalid');
			$i.closest('div').append($('<div class="verr">').text(err.message));
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

			var $c = $td.children('a, pre'), v = vs[k];
			if (typeof(v) == 'undefined') {
				v = '';
			}
			if (v && k.endsWith("_at")) {
				v = main[$td.hasClass('date') ? 'format_date' : 'format_time'](v);
			}
			($c.length ? $c : $td).text(v);
		}
	},
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

	init: function() {
		// sidenavi
		$('#sidenavi i').each(function() {
			$(this).attr('title', $(this).next('span').text());
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

	// load init
	$(window).on('load', main.init);
})(jQuery);

