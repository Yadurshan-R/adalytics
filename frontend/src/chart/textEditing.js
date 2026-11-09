// textEditing is the Text tool's box, which the drawing library leaves to the
// app: it opens where a text was placed or double-clicked, writes the typed
// text on the drawing, and removes a new text left empty or cancelled.
// textBox is the box's state ({ id, x, y, value, fresh } or null); getDm()
// is the drawing library's manager on screen.
export function textEditing(textBox, getDm) {
  // onTextEdit opens the text box where a text was placed or double-clicked.
  function onTextEdit(d, at) {
    textBox.value = { id: d.id, x: at.x, y: at.y, value: d.text ?? '', fresh: !d.text }
  }

  // submitText writes the typed text on the drawing; an empty new text is removed.
  function submitText(typed) {
    const box = textBox.value
    textBox.value = null
    const d = box && getDm()?.get(box.id)
    if (!d) return
    const text = String(typed ?? '').trim().slice(0, 80)
    if (text) getDm().update({ ...d, text })
    else if (box.fresh) getDm().remove(d.id)
  }

  // cancelText closes the text box (Esc); a new text is removed.
  function cancelText() {
    const box = textBox.value
    textBox.value = null
    if (box?.fresh && getDm()?.get(box.id)) getDm().remove(box.id)
  }

  return { onTextEdit, submitText, cancelText }
}
