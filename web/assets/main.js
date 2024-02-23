//------------------------------------------------------
var site = {
	base: '',
	body: 'body',
	static: '/static',
	cookie: { expires: 180 }
};

//------------------------------------------------------
function s_meta_props() {
	var m = {};
	$('meta').each(function() {
		var $t = $(this), a = $t.attr('property');
		if (a && a.substring(0, 2) == 's:') {
			m[a.substring(2)] = $t.attr('content');
		}
	});
	return m;
}

function s_safe_parse_json(s) {
	try {
		return $.parseJSON(s);
	} catch (e) {
		console.log(e);
		return s;
	}
}

function s_ajaf_error(data) {
	data = s_safe_parse_json(data);
	if (data && data.error) {
		$.toast({
			icon: 'error',
			text: data.error
		});
		return true;
	}
	return false;
}

function s_ajax_error(xhr, status, err) {
	err = err || status || 'Server error';
	if (xhr && xhr.responseJSON) {
		err = xhr.responseJSON.error || JSON.stringify(xhr.responseJSON, null, 4) || err;
	}

	$.toast({
		icon: 'error',
		text: err
	});
}

// popup messagebox
function s_popup_confirm(ps) {
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
					+ '<button class="btn btn-default cancel" popup-dismiss="true"><i class="fas fa-times"></i> <span>Cancel</span></button>'
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
}


//------------------------------------------------------
$(function() {
	// set cookie defaults
	$.extend($.cookie.defaults, site.cookie);

	// enable script cache
	$.enableScriptCache();
	
	$('[data-toggle=offcanvas]').click(function() {
		$('.row-offcanvas').toggleClass('active');
	});
	$('[data-toggle=tooltip]').tooltip();
	$('[data-toggle=popover]').popover();

	$('#sidenavi i').each(function() {
		$(this).attr('title', $(this).next('span').text());
	})

	$.extend(site, s_meta_props());

	$.extend($.toast.defaults, {
		position: 'top center'
	});

	$.extend($.popup.defaults, {
		transition: 'zoomIn'
	});
});

