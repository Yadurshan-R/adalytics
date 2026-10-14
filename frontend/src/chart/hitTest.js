// hitTest finds the drawing under the mouse, and which part of it: one of its
// points ('p1', 'p2'), the channel's width handle ('offset'), or the drawing
// itself ('body'). `layer` turns { time, price } points into pixels.

const NEAR = 6 // pixels: how close the mouse must be to a line or a point

// hitDrawing returns { index, part } for the topmost drawing at (x, y), or null.
export function hitDrawing(layer, x, y, size) {
  for (let i = layer.drawings.length - 1; i >= 0; i--) {
    const part = hitOne(layer, layer.drawings[i], x, y, size)
    if (part) return { index: i, part }
  }
  return null
}

// hitOne returns the part of drawing d at (x, y), or null.
function hitOne(layer, d, x, y, size) {
  const a = layer.xy(d.p1)
  const b = d.p2 ? layer.xy(d.p2) : null
  if (!a) return null
  const m = { x, y }
  if (b && near(m, b)) return 'p2'
  if (b && near(m, a)) return 'p1'
  switch (d.type) {
    case 'trend':
      return toSegment(m, a, b) <= NEAR ? 'body' : null
    case 'ray':
    case 'extended': {
      const from = d.type === 'extended' ? far(b, a, size) : a
      return toSegment(m, from, far(a, b, size)) <= NEAR ? 'body' : null
    }
    case 'hline':
      return Math.abs(y - a.y) <= NEAR ? 'body' : null
    case 'vline':
      return Math.abs(x - a.x) <= NEAR ? 'body' : null
    case 'rect':
    case 'fib':
      return inBox(m, a, b) ? 'body' : null
    case 'channel': {
      const c = layer.xy({ time: d.p1.time, price: d.p1.price + (d.offset || 0) })
      const e = layer.xy({ time: d.p2.time, price: d.p2.price + (d.offset || 0) })
      if (!c || !e) return null
      if (near(m, c) || near(m, e)) return 'offset'
      return toSegment(m, a, b) <= NEAR || toSegment(m, c, e) <= NEAR || inQuad(m, a, b, e, c) ? 'body' : null
    }
    case 'text': {
      const w = d.text.length * 8.5 + 4
      return x >= a.x - 4 && x <= a.x + w && Math.abs(y - a.y) <= 11 ? 'body' : null
    }
  }
  return null
}

const near = (m, p) => Math.hypot(m.x - p.x, m.y - p.y) <= NEAR + 2

// toSegment is the distance from point m to the line from a to b, in pixels.
function toSegment(m, a, b) {
  const dx = b.x - a.x
  const dy = b.y - a.y
  const len = dx * dx + dy * dy
  const t = len ? Math.max(0, Math.min(1, ((m.x - a.x) * dx + (m.y - a.y) * dy) / len)) : 0
  return Math.hypot(m.x - (a.x + t * dx), m.y - (a.y + t * dy))
}

// far carries the line from a through b well past the edge of the chart.
function far(a, b, size) {
  const dx = b.x - a.x
  const dy = b.y - a.y
  if (!dx && !dy) return b
  const k = (Math.max(size.width, size.height) * 4) / Math.hypot(dx, dy)
  return { x: b.x + dx * k, y: b.y + dy * k }
}

const inBox = (m, a, b) => m.x >= Math.min(a.x, b.x) && m.x <= Math.max(a.x, b.x) && m.y >= Math.min(a.y, b.y) && m.y <= Math.max(a.y, b.y)

// inQuad reports whether m is inside the four-sided shape p (a channel).
function inQuad(m, ...p) {
  let inside = false
  for (let i = 0, j = p.length - 1; i < p.length; j = i++) {
    if ((p[i].y > m.y) !== (p[j].y > m.y) && m.x < ((p[j].x - p[i].x) * (m.y - p[i].y)) / (p[j].y - p[i].y) + p[i].x) inside = !inside
  }
  return inside
}
