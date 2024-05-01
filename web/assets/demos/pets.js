$(function() {
	var sskey = "d.pets";

	//----------------------------------------------------
	// list: pager & sorter
	//
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
			beforeSend: function() {
				main.form_clear_invalid($f);
				main.loadmask();
			},
			success: function(data, ts, xhr) {
				$('#pets_list').html(data);

				$('#pets_list [checkall]').checkall();
				$('#pets_list [data-spy="pager"]').pager();
				$('#pets_list [data-spy="sortable"]').sortable();
			},
			error: main.form_ajax_error($f),
			complete: main.unloadmask
		});
		return false;
	}

	if (!location.search) {
		$('#pets_listform').formValues(main.ssload(sskey), true);
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


	//----------------------------------------------------
	// export
	//
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

	$('#pets_export').on('click', pets_export);


	//----------------------------------------------------
	// new
	//
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
	// detail
	//
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

	function pet_detail_submit() {
		$('#pet_detail_id').val() == '0' ? pet_create() : pet_update();
		return false;
	}

	$('#pets_list').on('click', 'button.view', function(evt) { return pet_detail(this, false); });
	$('#pets_list').on('click', 'button.edit', function(evt) { return pet_detail(this, true); });

	$('#pets_detail_popup')
		.on('loaded.popup', function() {
			$('#pets_detail_popup')
				.find('form').on('submit', pet_detail_submit).end()
				.find('.ui-popup-footer button[type=submit]').on('click', pet_detail_submit);
		}).on('shown.popup', function() {
			$('#pets_detail_popup')
				.find('.ui-popup-body').prop('scrollTop', 0).end()
				.find('[data-spy="niceSelect"]').niceSelect().end()
				.find('input[type="text"]').textclear().end()
				.find('textarea').autosize().textclear().enterfire();
			$(window).trigger('resize');
		});


	//----------------------------------------------------
	// update
	//
	var PGM = $('#pet_maps').data('gender');
	var POM = $('#pet_maps').data('origin');
	var PTM = $('#pet_maps').data('temper');
	var PHM = $('#pet_maps').data('habits');

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

				var pet = data.pet, $tr = $('tr#pet_' + pet.id);

				pet_set_tr_values($tr, pet);
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}


	//----------------------------------------------------
	// create
	//
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

				var pet = data.pet;
				var $tb = $('#pets_table > tbody'), $tr = $tb.children('tr.template').clone();

				$tr.attr({ 'class': '', 'id': 'pet_' + pet.id});
				$tr.find('td.check').append($('<input type="checkbox"/>').val(pet.id));
				$tb.prepend($tr);

				pet_set_tr_values($tr, pet);
				$tr.find('td.id, td.created_at').addClass('ro');
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}


	//----------------------------------------------------
	// deletes (selected / all)
	//
	function pets_deletes(all) {
		var $p = $(all ? '#pets_deleteall_popup' : '#pets_deletesel_popup').popup('update', { keyboard: false });
		var ids = all ? '*' : main.get_table_checked_ids($('#pets_table')).join(',');

		$.ajax({
			url: './deletes',
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

				(all ? pets_reset : pets_search)();
			},
			error: main.ajax_error,
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}

	$('#pets_deletesel_popup form').on('submit', function() { return pets_deletes(false); });
	$('#pets_deleteall_popup form').on('submit', function() { return pets_deletes(true); });


	//----------------------------------------------------
	// updates (selected / all)
	//
	function pets_updates() {
		var $p = $('#pets_bulkedit_popup').popup('update', { keyboard: false });

		$.ajax({
			url: './updates',
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

				if ($p.find('[name=id]').val() == '*') {
					pets_search();
					return;
				}

				var us = data.updates;
				if (us) {
					var ids = main.get_table_checked_ids($('#pets_table'));
					var $trs = main.get_table_trs('#pet_', ids);

					if (us.gender) {
						$trs.find('td.gender').text(PGM[us.gender]);
					}
					if (us.uborn_at) {
						$trs.find('td.born_at').text(main.format_date(us.born_at));
					}
					if (us.origin) {
						$trs.find('td.origin').text(POM[us.origin]);
					}
					if (us.temper) {
						$trs.find('td.temper').text(PTM[us.temper]);
					}
					if (us.uhabits) {
						var hs = [];
						$.each(us.habits, function(i, h) {
							hs.push($('<b>').text(PHM[h]));
						})
						$trs.find('td.habits').empty().append(hs);
					}
					main.blink($trs);
				}
			},
			error: main.form_ajax_error($p),
			complete:  function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}

	$('#pets_editsel').on('click', function() {
		var ids = main.get_table_checked_ids($('#pets_table'));
		$('#pets_bulkedit_popup')
			.find('.editsel').show().end()
			.find('.editall').hide().end()
			.find('input[name=id]').val(ids.join(',')).end()
			.popup('show');
	});
	$('#pets_editall').on('click', function() {
		$('#pets_bulkedit_popup')
			.find('.editsel').hide().end()
			.find('.editall').show().end()
			.find('input[name=id]').val('*').end()
			.popup('show');
	});
	$('#pets_bulkedit_popup')
		.find('.col-form-label > input').on('change', function() {
			var $t = $(this), c = $t.prop('checked');
			var $i = $t.parent().next().find(':input').prop('disabled', !c);
			if ($t.data('niceselect')) {
				$i.niceSelect('update');
			}
		}).end()
		.find('form').on('submit', pets_updates).end()
		.find('.ui-popup-footer button[type=submit]').on('click', pets_updates);
});
