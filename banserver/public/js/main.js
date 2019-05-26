$('#js-unlist').click((e) => {
  e.preventDefault();
  $.get(`/unlist/${$('#js-peerid').val()}`, () => {
    window.location.reload(true);
  });
});

$('#js-list').click((e) => {
    e.preventDefault();
    $.get(`/list/${$('#js-peerid').val()}`, () => {
        window.location.reload(true);
    });
});


$('#js-verify').click((e) => {
    e.preventDefault();
    $.get(`/verify/${$('#js-modid').val()}`, () => {
        window.location.reload(true);
    });
});

$('#js-unverify').click((e) => {
    e.preventDefault();
    $.get(`/unverify/${$('#js-modid').val()}`, () => {
        window.location.reload(true);
    });
});
