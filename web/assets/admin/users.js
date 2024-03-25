$(function() {
	var sskey = "a.users";

	//----------------------------------------------------
	// list: pager & sorter
	//----------------------------------------------------
	function users_reload() {
		$('#users_listform [name="p"]').val(1);
		users_reset();
	}

	function users_reset() {
		$('#users_listform').formClear().submit();
		return false;
	}

	function users_search() {
		$('body').loadmask();

		xdemo.sssave(sskey, xdemo.input_values($('#users_listform')));

		$.ajax({
			url: './list',
			type: 'POST',
			data: $('#users_listform').serialize(),
			success: function(data, ts, xhr) {
				$('#users_list').html(data);

				$('#users_list [checkall]').checkall();
				$('#users_list [data-spy="pager"]').pager();
				$('#users_list [data-spy="sortable"]').sortable();
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('body').unloadmask();
			}
		});
		return false;
	}

	function users_export() {
		$('body').loadmask();

		$.ajaf({
			url: './export/csv',
			type: 'POST',
			data: $('#users_listform').serializeArray(),
			error: xdemo.ajax_error,
			complete: function() {
				$('body').unloadmask();
			}
		});
		return false;
	}

	function has_search_params() {
		return $('users_listform input:checked').length || $('#users_listform input[type=text]').filter(function() { return $(this).val(); }).length;
	}

	if (!location.search) {
		$('#users_listform').formValues(xdemo.ssload(sskey));
	}
	if (has_search_params()) {
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
	function user_detail(evt) {
		evt.stopPropagation();

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
	function user_new(evt) {
		evt.stopPropagation();

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
		$('#users_detail_popup').loadmask();
		$.ajax({
			url: './update',
			type: 'POST',
			data: $('#users_detail_popup form').serialize(),
			dataType: 'json',
			success: function(data, ts, xhr) {
				$('#users_detail_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var usr = data.result, $tr = $('tr#usr_' + usr.id);

				xdemo.set_table_tr_values($tr, usr);
				$tr.attr('class', '').addClass(usr.status);
				$tr.find('td.status').text(USM[usr.status]);
				$tr.find('td.role').text(URM[usr.role]);
				xdemo.blink($tr);
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('#users_detail_popup').unloadmask();
			}
		});
		return false;
	}

	//----------------------------------------------------
	// create
	//----------------------------------------------------
	function user_create() {
		$('#users_detail_popup').loadmask();
		$.ajax({
			url: './create',
			type: 'POST',
			data: $('#users_detail_popup form').serialize(),
			dataType: 'json',
			success: function(data, ts, xhr) {
				$('#users_detail_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var usr = data.result;
				var $tr = $('#usr_new').clone();

				$tr.removeClass('hidden').attr('id', 'usr_' + usr.id);
				$tr.find('td.check').append($('<input type="checkbox"/>').val(usr.id));
				$('#users_table > tbody').prepend($tr);

				xdemo.set_table_tr_values($tr, usr);
				$tr.addClass(usr.status);
				$tr.find('td.status').text(USM[usr.status]);
				$tr.find('td.role').text(URM[usr.role]);
				$tr.find('td.id, td.created_at').addClass('ro');
				xdemo.blink($tr);
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('#users_detail_popup').unloadmask();
			}
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
		var ids = xdemo.get_table_checked_ids($('#users_table'));

		$('#users_delete_popup').loadmask();
		$.ajax({
			url: './delete',
			type: 'POST',
			data: {
				_token_: xdemo.token,
				id: ids
			},
			dataType: 'json',
			success: function(data, ts, xhr) {
				$('#users_delete_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				users_search();
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('#users_delete_popup').unloadmask();
			}
		});
		return false;
	}

	$('#users_delete_popup form').submit(users_delete);


	//----------------------------------------------------
	// clear
	//----------------------------------------------------
	function users_clear() {
		$('#users_clear_popup').loadmask();
		$.ajax({
			url: './clear',
			type: 'POST',
			data: {
				_token_: xdemo.token
			},
			dataType: 'json',
			success: function(data, ts, xhr) {
				$('#users_clear_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				users_reload();
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('#users_clear_popup').unloadmask();
			}
		});
		return false;
	}

	$('#users_clear_popup form').submit(users_clear);

	//----------------------------------------------------
	// users enable/disable
	//----------------------------------------------------
	function users_enable(en) {
		var $p = $(en ? '#users_enable_popup' : '#users_disable_popup');
		var ids = xdemo.get_table_checked_ids($('#users_table'));

		$p.loadmask();
		$.ajax({
			url: en ? 'enable' : 'disable',
			type: 'POST',
			data: {
				_token_: xdemo.token,
				id: ids
			},
			dataType: 'json',
			success: function(data, ts, xhr) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var sts = en ? 'A' : 'D';
				var $trs = xdemo.get_table_trs('#usr_', ids);
				$trs.attr('class', '').addClass(sts);
				$trs.find('td.status').text(USM[sts]);
				xdemo.blink($trs);
			},
			error: xdemo.ajax_error,
			complete: function() {
				$p.unloadmask();
			}
		});
		return false;
	}

	$('#users_enable_popup form').submit(function() { return users_enable(true); });
	$('#users_disable_popup form').submit(function() { return users_enable(false); });
});
