$(function() {
	//----------------------------------------------------
	// pager & sorter
	//----------------------------------------------------
	$('.ui-pager').on('goto.pager', function(evt, pno) {
		$('#users_pageform').find('input[name="p"]').val(pno).end().submit();
	});

	$('.ui-pager select').on('change', function() {
		$('#users_pageform').find('input[name="l"]').val($(this).val()).end().submit();
	});

	$('#users_table').on('sort.sortable', function(evt, col, dir) {
		$('#users_pageform')
			.find('input[name="c"]').val(col).end()
			.find('input[name="d"]').val(dir).end()
			.submit();
	});


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

	$('#users_table').on('click', 'button.edit', user_detail);

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
				xdemo.blink($tr);
				$tr.attr('class', '').addClass(usr.status);
				$tr.find('td.status').text(USM[usr.status]);
				$tr.find('td.role').text(URM[usr.role]);
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
				xdemo.blink($tr);
				$tr.addClass(usr.status);
				$tr.find('td.status').text(USM[usr.status]);
				$tr.find('td.role').text(URM[usr.role]);
				$tr.find('td.id, td.created_at').addClass('ro');
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
					text: data.success,
					hideAfter: 3000,
					afterHidden: function() {
						location.reload();
					}
				});
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
				_token_: xdemo.token,
				k: $('#users_pageform').find('input[name="k"]').val()
			},
			dataType: 'json',
			success: function(data, ts, xhr) {
				$('#users_clear_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success,
					hideAfter: 3000,
					afterHidden: function() {
						location.reload();
					}
				});
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
	// users enable
	//----------------------------------------------------
	function users_enable() {
		var ids = xdemo.get_table_checked_ids($('#users_table'));

		$('#users_enable_popup').loadmask();
		$.ajax({
			url: './enable',
			type: 'POST',
			data: {
				_token_: xdemo.token,
				id: ids
			},
			dataType: 'json',
			success: function(data, ts, xhr) {
				$('#users_enable_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success,
					hideAfter: 3000,
					afterHidden: function() {
						location.reload();
					}
				});
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('#users_enable_popup').unloadmask();
			}
		});
		return false;
	}

	$('#users_enable_popup form').submit(users_enable);

	//----------------------------------------------------
	// disable
	//----------------------------------------------------
	function users_disable() {
		var ids = xdemo.get_table_checked_ids($('#users_table'));

		$('#users_disable_popup').loadmask();
		$.ajax({
			url: './disable',
			type: 'POST',
			data: {
				_token_: xdemo.token,
				id: ids
			},
			dataType: 'json',
			success: function(data, ts, xhr) {
				$('#users_disable_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success,
					hideAfter: 3000,
					afterHidden: function() {
						location.reload();
					}
				});
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('#users_disable_popup').unloadmask();
			}
		});
		return false;
	}

	$('#users_disable_popup form').submit(users_disable);
});
