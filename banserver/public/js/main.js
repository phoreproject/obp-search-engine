$('#js-unlist').click((e) => {
  e.preventDefault()
  $.get(`/unlist/${$('#js-peerid').val()}`, () => {
    window.location.reload(true)
  })
})