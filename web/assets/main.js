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
		err = err || status || 'Server error';
		if (xhr && xhr.responseJSON) {
			err = xhr.responseJSON.error || JSON.stringify(xhr.responseJSON, null, 4) || err;
		}

		if (err.title && err.detail) {
			$.toast({
				icon: 'error',
				heading: err.title,
				text: err.detail
			});
			return;
		}

		$.toast({
			icon: 'error',
			text: err
		});
	},

	// load mask
	loadmask: function() {
		$('body').loadmask();
	},
	unloadmask: function() {
		$('body').unloadmask();
	},

	// table
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

		// blink tr
		xdemo.blink($tr);
	},
	blink: function($e) {
		$e.addClass('ui-blink-1s2');
		setTimeout(function() { $e.removeClass('ui-blink-1s2'); }, 2000);
	},

	// popup messagebox
	popup_confirm: function(ps) {
		var $pc = $('#s_popup_confirm');
		if (!$pc.length) {
			$pc = $('<div id="s_popup_confirm" class="ui-popup s-popup-confirm" popup-mask="true" popup-position="center" popup-closer="false">'
					+ '<h5 class="ui-popup-header"></h5>'
					+ '<div class="ui-popup-body">'
						+ '<i class="icon fas fa-3x fa-question-circle"></i>'
						+ '<div class="msg"></div>'
					+ '</div>'
					+ '<div class="ui-popup-footer">'
						+ '<button class="btn btn-primary ok"><i class="fas fa-check"></i> <span>OK</span></button>\n'
						+ '<button class="btn btn-secondary cancel" popup-dismiss="true"><i class="fas fa-times"></i> <span>Cancel</span></button>'
					+ '</div>'
				+ '</div>'
			);
			$pc.popup();
		}

		$pc.find('.ui-popup-header').text(ps.title);
		$pc.find('.msg').text(ps.message);
		if (ps.icon) {
			if (ps.icon.ok) {
				$pc.find('.ok>i').prop('class', ps.icon.ok);
			}
			if (ps.icon.cancel) {
				$pc.find('.cancel>i').prop('class', ps.icon.cancel);
			}
		}
		if (ps.text) {
			if (ps.text.ok) {
				$pc.find('.ok>span').text(ps.text.ok);
			}
			if (ps.text.cancel) {
				$pc.find('.cancel>span').text(ps.text.cancel);
			}
		}

		$pc.find('.ok').off('click').on('click', function() {
			$pc.popup('hide');
			ps.onok();
		})
		$pc.popup('show');
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

