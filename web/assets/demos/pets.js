$(function() {
	var sskey = "d.pets";

	//----------------------------------------------------
	// list: pager & sorter
	//----------------------------------------------------
	function pets_reset() {
		$('#pets_listform [name="p"]').val(1);
		$('#pets_listform').formClear(true).submit();
		return false;
	}

	function pets_search() {
		var $f = $('#pets_listform');

		main.sssave(sskey, main.form_input_values($f));

		$.ajax({
			url: './list',
			type: 'POST',
			data: $f.serialize(),
			beforeSend: main.loadmask,
			success: function(data, ts, xhr) {
				$('#pets_list').html(data);

				$('#pets_list [checkall]').checkall();
				$('#pets_list [data-spy="pager"]').pager();
				$('#pets_list [data-spy="sortable"]').sortable();
			},
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}

	function pets_export() {
		$.ajaf({
			url: './export/csv',
			type: 'POST',
			data: $('#pets_listform').serializeArray(),
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}

	if (!location.search) {
		$('#pets_listform').formValues(main.ssload(sskey));
	}
	if (main.form_has_inputs($('#pets_listform'))) {
		$('#pets_listfset').fieldset('expand', 'show');
	}

	$('#pets_list')
		.on('goto.pager', '.ui-pager', function(evt, pno) {
			$('#pets_listform').find('input[name="p"]').val(pno).end().submit();
		})
		.on('change', '.ui-pager select', function() {
			$('#pets_listform').find('input[name="l"]').val($(this).val()).end().submit();
		})
		.on('sort.sortable', '#pets_table', function(evt, col, dir) {
			$('#pets_listform')
				.find('input[name="c"]').val(col).end()
				.find('input[name="d"]').val(dir).end()
				.submit();
		});

	$('#pets_listform').on('submit', pets_search).submit();
	$('#pets_listform').on('reset', pets_reset);
	$('#pets_export').on('click', pets_export);


	//----------------------------------------------------
	// detail
	//----------------------------------------------------
	function pet_detail(self, edit) {
		var $tr = $(self).closest('tr');
		var params = {
			id: $tr.attr('id').replace('pet_', '')
		};

		$('#pets_detail_popup').popup({
			loaded: false,
			keyboard: !edit,
			ajax: {
				url: edit ? "edit" : "view",
				method: 'GET',
				data: params
			}
		}).popup('show', this);

		return false;
	}

	$('#pets_list').on('click', 'button.view', function(evt) { return pet_detail(this, false); });
	$('#pets_list').on('click', 'button.edit', function(evt) { return pet_detail(this, true); });

	//----------------------------------------------------
	// new
	//----------------------------------------------------
	function pet_new() {
		$('#pets_detail_popup').popup({
			loaded: false,
			keyboard: false,
			ajax: {
				url: './new',
				method: 'GET'
			}
		}).popup('show', this);

		return false;
	}

	$('#pets_new').on('click', pet_new);

	//----------------------------------------------------
	// update
	//----------------------------------------------------
	var PGM = $('#pet_maps').data('gender');
	var POM = $('#pet_maps').data('origin');
	var PTM = $('#pet_maps').data('temper');
	var PHM = $('#pet_maps').data('habits');

	function pet_update() {
		var $p = $('#pets_detail_popup');

		$.ajax({
			url: './update',
			type: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
				$('#pets_detail_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var pet = data.result, $tr = $('tr#pet_' + pet.id);
				pet_set_tr_values($tr, pet);
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	function pet_set_tr_values($tr, pet) {
		main.set_table_tr_values($tr, pet);
		$tr.find('td.gender').text(PGM[pet.gender]);
		$tr.find('td.origin').text(POM[pet.origin]);
		$tr.find('td.temper').text(PTM[pet.temper]);
		var hs = [];
		$.each(pet.habits, function(i, h) {
			hs.push($('<b>').text(PHM[h]));
		})
		$tr.find('td.habits').empty().append(hs);
		main.blink($tr);
	}

	//----------------------------------------------------
	// create
	//----------------------------------------------------
	function pet_create() {
		var $p = $('#pets_detail_popup');

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

				var pet = data.result;
				var $tr = $('#pet_new').clone();

				$tr.removeClass('hidden').attr('id', 'pet_' + pet.id);
				$tr.find('td.check').append($('<input type="checkbox"/>').val(pet.id));
				$('#pets_table > tbody').prepend($tr);

				pet_set_tr_values($tr, pet);
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
	function pet_detail_submit() {
		$('#pet_detail_id').val() == '0' ? pet_create() : pet_update();
		return false;
	}

	$('#pets_detail_popup')
		.on('loaded.popup', function() {
			$('#pets_detail_popup')
				.find('form').submit(pet_detail_submit).end()
				.find('.ui-popup-footer button[type=submit]').click(pet_detail_submit);
		}).on('shown.popup', function() {
			$('#pets_detail_popup')
				.find('.ui-popup-body').prop('scrollTop', 0).end()
				.find('[data-spy="niceSelect"]').niceSelect().end()
				.find('input[type="text"]').textclear().end()
				.find('textarea').autosize().textclear().enterfire();
			$(window).trigger('resize');
		});

	//----------------------------------------------------
	// delete
	//----------------------------------------------------
	function pets_delete() {
		var $p = $('#pets_delete_popup');
		var ids = main.get_table_checked_ids($('#pets_table'));

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

				pets_search();
			},
			error: main.ajax_error,
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	$('#pets_delete_popup form').submit(pets_delete);


	//----------------------------------------------------
	// clear
	//----------------------------------------------------
	function pets_clear() {
		var $p = $('#pets_clear_popup');
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

				pets_reset();
			},
			error: main.ajax_error,
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	$('#pets_clear_popup form').submit(pets_clear);
});
