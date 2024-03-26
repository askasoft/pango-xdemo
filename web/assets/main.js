var xdemo = {
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
		return xdemo.fmt_date.format(d);
	},
	format_time: function(d) {
		if (typeof(d) == 'string') {
			d = new Date(d);
		}
		return xdemo.fmt_time.format(d);
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

	// ajaf error handler
	ajaf_error: function(data) {
		data = xdemo.safe_parse_json(data);
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
	ajax_error: function(xhr, status, err) {
		if (xhr.readyState == 0) { // window unload
			console.log('ajax canceled.');
			return;
		}

		var afterHidden;
		if (xhr.status == 401) { // unauthorized
			afterHidden = function() {
				window.location.href = xdemo.base + '/login/';
			};
		}

		err = err || status || 'Server error';
		if (xhr.responseJSON) {
			err = xhr.responseJSON.error || JSON.stringify(xhr.responseJSON, null, 4) || err;
		}

		if (err.title && err.detail) {
			$.toast({
				icon: 'error',
				heading: err.title,
				text: err.detail,
				afterHidden: afterHidden
			});
			return;
		}

		$.toast({
			icon: 'error',
			text: err,
			afterHidden: afterHidden
		});
	},

	// load mask
	loadmask: function() {
		$('body').loadmask();
	},
	unloadmask: function() {
		$('body').unloadmask();
	},

	// form input (not hidden) values
	input_values: function($f) {
		var vs = $f.formValues();
		delete vs._token_;
		return vs;
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

			var $c = $td.children('a, pre'), v = vs[k] || '';
			if (v && k.endsWith("_at")) {
				v = xdemo.format_time(v);
			}
			($c.length ? $c : $td).text(v);
		}
	},
	blink: function($e) {
		$e.addClass('ui-blink-1s2');
		setTimeout(function() { $e.removeClass('ui-blink-1s2'); }, 2000);
	},

	init: function() {
		// set cookie defaults
		$.extend($.cookie.defaults, xdemo.cookie);

		// enable script cache
		$.enableScriptCache();
		
		// get meta properties
		$.extend(xdemo, xdemo.meta_props());

		// set plugins defaults
		$.extend($.toast.defaults, {
			position: 'top center'
		});

		$.extend($.popup.defaults, {
			transition: 'zoomIn'
		});

		// enable bootstrap UI
		$('[data-toggle=offcanvas]').click(function() {
			$('.row-offcanvas').toggleClass('active');
		});
		$('[data-toggle=tooltip]').tooltip();
		$('[data-toggle=popover]').popover();

		// sidenavi
		$('#sidenavi i').each(function() {
			$(this).attr('title', $(this).next('span').text());
		})

		// header theme switch
		$('#header li.theme a').click(function() {
			var $a = $(this), t = $a.attr('href').substring(1);
			$('body').attr('data-bs-theme', t);
			localStorage.theme = t;
			$('#header li.theme a').removeClass('active');
			$a.addClass('active');
			return false;
		}).filter('[href="#' + localStorage.theme + '"]').trigger('click');
	}
};

//------------------------------------------------------
$(function() {
	xdemo.init();
});

