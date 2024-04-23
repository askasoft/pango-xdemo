$(function() {
	var sskey = "a.users";

	//----------------------------------------------------
	// list: pager & sorter
	//----------------------------------------------------
	function users_reset() {
		$('#users_listform [name="p"]').val(1);
		$('#users_listform').formClear(true).submit();
		return false;
	}

	function users_search() {
		var $f = $('#users_listform');

		main.sssave(sskey, main.form_input_values($f));

		$.ajax({
			url: './list',
			type: 'POST',
			data: $f.serialize(),
			beforeSend: main.loadmask,
			success: function(data, ts, xhr) {
				$('#users_list').html(data);

				$('#users_list [checkall]').checkall();
				$('#users_list [data-spy="pager"]').pager();
				$('#users_list [data-spy="sortable"]').sortable();
			},
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}

	function users_export() {
		$.ajaf({
			url: './export/csv',
			type: 'POST',
			data: $('#users_listform').serializeArray(),
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}

	if (!location.search) {
		$('#users_listform').formValues(main.ssload(sskey));
	}
	if (main.form_has_inputs($('#users_listform'))) {
		$('#users_listfset').fieldset('expand', 'show');
	}

	$('#users_list')
		.on('goto.pager', '.ui-pager', function(evt, pno) {
			$('#users_listform').find('input[name="p"]').val(pno).end().submit();
		})
		.on('change', '.ui-pager select', function() {
			$('#users_listform').find('input[name="l"]').val($(this).val()).end().submit();
		})
		.on('sort.sortable', '#users_table', function(evt, col, dir) {
			$('#users_listform')
				.find('input[name="c"]').val(col).end()
				.find('input[name="d"]').val(dir).end()
				.submit();
		});

	$('#users_listform').on('submit', users_search).submit();
	$('#users_listform').on('reset', users_reset);
	$('#users_export').on('click', users_export);


	//----------------------------------------------------
	// detail
	//----------------------------------------------------
	function user_detail() {
		var $tr = $(this).closest('tr');
		var params = {
			id: $tr.attr('id').replace('usr_', '')
		};

		$('#users_detail_popup').popup({
			loaded: false,
			ajax: {
				url: './detail',
				method: 'GET',
				data: params
			}
		}).popup('show', this);

		return false;
	}

	$('#users_list').on('click', 'button.edit', user_detail);

	//----------------------------------------------------
	// new
	//----------------------------------------------------
	function user_new() {
		$('#users_detail_popup').popup({
			loaded: false,
			ajax: {
				url: './new',
				method: 'GET'
			}
		}).popup('show', this);

		return false;
	}

	$('#users_new').on('click', user_new);

	//----------------------------------------------------
	// update
	//----------------------------------------------------
	var USM = $('#user_maps').data('status');
	var URM = $('#user_maps').data('role');

	function user_update() {
		var $p = $('#users_detail_popup');

		$.ajax({
			url: './update',
			type: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
				$('#users_detail_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var usr = data.result, $tr = $('tr#usr_' + usr.id);

				user_set_tr_values($tr, usr);
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	function user_set_tr_values($tr, usr) {
		main.set_table_tr_values($tr, usr);
		$tr.attr('class', '').addClass(usr.status);
		$tr.find('td.status').text(USM[usr.status]);
		$tr.find('td.role').text(URM[usr.role]);
		main.blink($tr);
	}

	//----------------------------------------------------
	// create
	//----------------------------------------------------
	function user_create() {
		var $p = $('#users_detail_popup');

		$.ajax({
			url: './create',
			type: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var usr = data.result;
				var $tb = $('#users_table > tbody'), $tr = $tb.children('tr.template').clone();

				$tr.attr({ 'class': '', 'id': 'usr_' + usr.id});
				$tr.find('td.check').append($('<input type="checkbox"/>').val(usr.id));
				$tb.prepend($tr);

				user_set_tr_values($tr, usr);
				$tr.find('td.id, td.created_at').addClass('ro');
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	//----------------------------------------------------
	// detail popup
	//----------------------------------------------------
	function user_detail_submit() {
		$('#user_detail_id').val() == '0' ? user_create() : user_update();
		return false;
	}

	$('#users_detail_popup')
		.on('loaded.popup', function() {
			$('#users_detail_popup')
				.find('form').submit(user_detail_submit).end()
				.find('.ui-popup-footer button[type=submit]').click(user_detail_submit);
		}).on('shown.popup', function() {
			$('#users_detail_popup')
				.find('.ui-popup-body').prop('scrollTop', 0).end()
				.find('input[type="text"]').textclear().end()
				.find('textarea').autosize().textclear().enterfire();
			$(window).trigger('resize');
		});

	//----------------------------------------------------
	// delete
	//----------------------------------------------------
	function users_delete() {
		var $p = $('#users_delete_popup');
		var ids = main.get_table_checked_ids($('#users_table'));

		$.ajax({
			url: './delete',
			type: 'POST',
			data: {
				_token_: main.token,
				id: ids
			},
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				users_search();
			},
			error: main.ajax_error,
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	$('#users_delete_popup form').submit(users_delete);


	//----------------------------------------------------
	// clear
	//----------------------------------------------------
	function users_clear() {
		var $p = $('#users_clear_popup');
		$.ajax({
			url: './clear',
			type: 'POST',
			data: {
				_token_: main.token
			},
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				users_reset();
			},
			error: main.ajax_error,
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	$('#users_clear_popup form').submit(users_clear);

	//----------------------------------------------------
	// users enable/disable
	//----------------------------------------------------
	function users_enable(en) {
		var $p = $(en ? '#users_enable_popup' : '#users_disable_popup');
		var ids = main.get_table_checked_ids($('#users_table'));

		$.ajax({
			url: en ? 'enable' : 'disable',
			type: 'POST',
			data: {
				_token_: main.token,
				id: ids
			},
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var sts = en ? 'A' : 'D';
				var $trs = main.get_table_trs('#usr_', ids);
				$trs.attr('class', '').addClass(sts);
				$trs.find('td.status').text(USM[sts]);
				main.blink($trs);
			},
			error: main.ajax_error,
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	$('#users_enable_popup form').submit(function() { return users_enable(true); });
	$('#users_disable_popup form').submit(function() { return users_enable(false); });
});
