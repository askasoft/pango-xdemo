(function($) {
	var sskey = "a.auditlogs";

	//----------------------------------------------------
	// list: pager & sorter
	//
	function auditlogs_reset() {
		$('#auditlogs_listform [name="p"]').val(1);
		$('#auditlogs_listform').formClear(true).submit();
		return false;
	}

	function auditlogs_search(evt, callback) {
		var $f = $('#auditlogs_listform'), vs = main.form_input_values($f);

		main.sssave(sskey, vs);
		main.location_replace_search(vs);

		$.ajax({
			url: './list',
			method: 'POST',
			data: $.param(vs, true),
			beforeSend: function() {
				main.form_clear_invalid($f);
				main.loadmask();
			},
			success: main.list_builder($('#auditlogs_list'), callback),
			error: main.form_ajax_error($f),
			complete: main.unloadmask
		});
		return false;
	}


	//----------------------------------------------------
	// export
	//
	function auditlogs_export() {
		$.ajaf({
			url: './export/csv',
			method: 'POST',
			data: $('#auditlogs_listform').serializeArray(),
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}


	//----------------------------------------------------
	// deletes (selected / all)
	//
	function auditlogs_deletes(all) {
		var $p = $(all ? '#auditlogs_deleteall_popup' : '#auditlogs_deletesel_popup').popup('update', { keyboard: false });
		var ids = all ? '*' : main.get_table_checked_ids($('#auditlogs_table')).join(',');

		$.ajax({
			url: './deletes',
			method: 'POST',
			data: {
				id: ids
			},
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				(all ? auditlogs_reset : auditlogs_search)();
			},
			error: main.ajax_error,
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}


	//----------------------------------------------------
	// deletes (batch)
	//
	function auditlogs_deletebat() {
		var $p = $('#auditlogs_deletebat_popup').popup('update', { keyboard: false });
		var vs = main.form_input_values($p.find('form'));

		$.ajax({
			url: './deleteb',
			method: 'POST',
			data: $.param(vs, true),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				auditlogs_search();
			},
			error: main.form_ajax_error($p),
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}

	function on_detail_click() {
		var $b = $('#auditlogs_detail_popup .ui-popup-body').empty();
		var detail = $(this).text();
		if (!detail) {
			return false;
		}

		try {
			var jo = JSON.parse(detail);
			var $t = $('<table class="table">'), $tb = $('<tbody>');
			for (var k in jo) {
				var v = jo[k];
				if (typeof(v) != 'string') {
					v = JSON.stringify(v);
				}
				$tb.append($('<tr>').append(
					$('<td>').text(k),
					$('<td>').append($('<pre>').text(v))
				));
			}
			$b.append($t.append($tb));
		} catch (ex) {
			$b.append($('<pre>').text(detail));
		}

		$('#auditlogs_detail_popup').popup('toggle', this);
		return false;
	}

	//----------------------------------------------------
	// init
	//
	function auditlogs_init() {
		main.list_init('auditlogs', sskey);

		$('#auditlogs_listform')
			.on('reset', auditlogs_reset)
			.on('submit', auditlogs_search)
			.submit();

		$('#auditlogs_export').on('click', auditlogs_export);

		$('#auditlogs_deletesel_popup form').on('submit', auditlogs_deletes.callback(false));
		$('#auditlogs_deleteall_popup form').on('submit', auditlogs_deletes.callback(true));

		$('#auditlogs_deletebat_popup')
			.on('submit', 'form', auditlogs_deletebat)
			.on('click', '.ui-popup-footer button[type=submit]', auditlogs_deletebat);

		$('#auditlogs_list').on('click', 'td.detail > pre', on_detail_click);
	}

	$(window).on('load', auditlogs_init);
})(jQuery);
