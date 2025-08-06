(function($) {
	var sskey = "d.files";

	//----------------------------------------------------
	// list: pager & sorter
	//
	function files_reset() {
		$('#files_listform [name="p"]').val(1);
		$('#files_listform').formClear(true).submit();
		return false;
	}

	function files_search(evt, callback) {
		var $f = $('#files_listform'), vs = main.form_input_values($f);

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
			success: main.list_builder($('#files_list'), callback),
			error: main.form_ajax_error($f),
			complete: main.unloadmask
		});
		return false;
	}

	function files_prev_page(callback) {
		var pno = $('#files_list > .ui-pager > .pagination > .page-item.prev > a').attr('pageno');
		$('#files_listform input[name="p"]').val(pno);
		files_search(null, callback);
	}

	function files_next_page(callback) {
		var pno = $('#files_list > .ui-pager > .pagination > .page-item.next > a').attr('pageno');
		$('#files_listform input[name="p"]').val(pno);
		files_search(null, callback);
	}


	//----------------------------------------------------
	// deletes (selected / all)
	//
	function files_deletes(all) {
		var $p = $(all ? '#files_deleteall_popup' : '#files_deletesel_popup').popup('update', { keyboard: false });
		var ids = all ? '*' : main.get_table_checked_ids($('#files_table')).join(',');

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

				(all ? files_reset : files_search)();
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
	function files_deletebat() {
		var $p = $('#files_deletebat_popup').popup('update', { keyboard: false });
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

				files_search();
			},
			error: main.form_ajax_error($p),
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}


	//----------------------------------------------------
	// init
	//
	function files_init() {
		main.list_init('files', sskey);
	
		$('#files_listform')
			.on('reset', files_reset)
			.on('submit', files_search)
			.submit();


		$('#files_deletesel_popup form').on('submit', files_deletes.callback(false));
		$('#files_deleteall_popup form').on('submit', files_deletes.callback(true));

		$('#files_deletebat_popup')
			.on('submit', 'form', files_deletebat)
			.on('click', '.ui-popup-footer button[type=submit]', files_deletebat);
	}

	$(window).on('load', files_init);
})(jQuery);
