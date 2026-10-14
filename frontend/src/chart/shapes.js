// shapes.js draws each kind of drawing on the chart's canvas: lines, rays,
// channels, rectangles, Fibonacci levels, text and the ruler. `layer` turns
// { time, price } points into pixels (see drawingLayer.js).
import { formatPrice } from '../format.js'

const BLUE = '#2962ff'
const BULL = '#089981'
const BEAR = '#f23645'
const FIB_LEVELS = [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1]

// drawShape draws one drawing. A drawing being placed or selected (marked)
// shows its handles.
export function drawShape(ctx, layer, d, size, marked) {
  const a = layer.xy(d.p1)
  const b = d.p2 ? layer.xy(d.p2) : null
  if (!a) return
  ctx.strokeStyle = BLUE
  ctx.fillStyle = BLUE
  ctx.lineWidth = 2
  switch (d.type) {
    case 'trend':
    case 'ray':
    case 'extended': {
      if (!b) return
      const from = d.type === 'extended' ? extend(b, a, size) : a
      const to = d.type === 'trend' ? b : extend(a, b, size)
      line(ctx, from.x, from.y, to.x, to.y)
      handle(ctx, a)
      handle(ctx, b)
      break
    }
    case 'hline':
      line(ctx, 0, a.y, size.width, a.y)
      label(ctx, formatPrice(d.p1.price), size.width - 6, a.y)
      if (marked) handle(ctx, { x: size.width / 2, y: a.y })
      break
    case 'vline':
      line(ctx, a.x, 0, a.x, size.height)
      if (marked) handle(ctx, { x: a.x, y: size.height / 2 })
      break
    case 'channel':
      if (b) channel(ctx, layer, d, a, b)
      break
    case 'rect': {
      if (!b) return
      const [x, y, w, h] = [Math.min(a.x, b.x), Math.min(a.y, b.y), Math.abs(b.x - a.x), Math.abs(b.y - a.y)]
      ctx.fillStyle = 'rgba(41, 98, 255, 0.10)'
      ctx.fillRect(x, y, w, h)
      ctx.strokeRect(x, y, w, h)
      if (marked) [a, b].forEach((p) => handle(ctx, p))
      break
    }
    case 'fib':
      if (!b) return
      fib(ctx, layer, d, a, b)
      if (marked) [a, b].forEach((p) => handle(ctx, p))
      break
    case 'text':
      ctx.font = '600 14px Inter, system-ui, sans-serif'
      ctx.textAlign = 'left'
      ctx.textBaseline = 'middle'
      ctx.lineWidth = 4
      ctx.strokeStyle = '#ffffff'
      ctx.strokeText(d.text, a.x, a.y)
      ctx.fillStyle = '#131722'
      ctx.fillText(d.text, a.x, a.y)
      if (marked) {
        ctx.lineWidth = 1
        ctx.strokeStyle = BLUE
        ctx.strokeRect(a.x - 4, a.y - 11, ctx.measureText(d.text).width + 8, 22)
      }
      break
  }
}

// channel draws two parallel lines, the space between them and a dashed
// middle line. d.offset is how far (in ADA) the second line is from the first.
function channel(ctx, layer, d, a, b) {
  const off = d.offset || 0
  const c = layer.xy({ time: d.p1.time, price: d.p1.price + off })
  const e = layer.xy({ time: d.p2.time, price: d.p2.price + off })
  if (!c || !e) return
  ctx.fillStyle = 'rgba(41, 98, 255, 0.10)'
  ctx.beginPath()
  ctx.moveTo(a.x, a.y)
  ctx.lineTo(b.x, b.y)
  ctx.lineTo(e.x, e.y)
  ctx.lineTo(c.x, c.y)
  ctx.closePath()
  ctx.fill()
  line(ctx, a.x, a.y, b.x, b.y)
  line(ctx, c.x, c.y, e.x, e.y)
  ctx.lineWidth = 1
  ctx.setLineDash([4, 4])
  line(ctx, (a.x + c.x) / 2, (a.y + c.y) / 2, (b.x + e.x) / 2, (b.y + e.y) / 2)
  ctx.setLineDash([])
  ctx.lineWidth = 2
  for (const p of [a, b, c, e]) handle(ctx, p)
}

// fib draws Fibonacci retracement levels between the two points.
function fib(ctx, layer, d, a, b) {
  const left = Math.min(a.x, b.x)
  const right = Math.max(a.x, b.x)
  ctx.lineWidth = 1
  ctx.textAlign = 'left'
  ctx.textBaseline = 'bottom'
  for (const level of FIB_LEVELS) {
    const price = d.p2.price + (d.p1.price - d.p2.price) * level
    const y = layer.y(price)
    if (y == null) continue
    line(ctx, left, y, right, y)
    ctx.fillText(`${level} (${formatPrice(price)})`, left + 4, y - 3)
  }
  ctx.setLineDash([4, 4])
  line(ctx, a.x, a.y, b.x, b.y)
  ctx.setLineDash([])
}

// drawRuler shows the price change and time between two points.
export function drawRuler(ctx, layer, r) {
  const a = layer.xy(r.p1)
  const b = layer.xy(r.p2)
  if (!a || !b) return
  const up = r.p2.price >= r.p1.price
  const col = up ? BULL : BEAR
  ctx.fillStyle = up ? 'rgba(8, 153, 129, 0.14)' : 'rgba(242, 54, 69, 0.14)'
  ctx.fillRect(Math.min(a.x, b.x), Math.min(a.y, b.y), Math.abs(b.x - a.x), Math.abs(b.y - a.y))
  ctx.strokeStyle = col
  ctx.lineWidth = 1.5
  line(ctx, a.x, a.y, b.x, b.y)

  const pct = r.p1.price > 0 ? (r.p2.price / r.p1.price - 1) * 100 : 0
  const diff = r.p2.price - r.p1.price
  const bars = Math.round((r.p2.time - r.p1.time) / layer.step)
  const lines = [
    `${pct >= 0 ? '+' : ''}${pct.toFixed(2)}%  (${diff >= 0 ? '+' : '−'}${formatPrice(Math.abs(diff))})`,
    `${Math.abs(bars)} bars, ${duration(Math.abs(r.p2.time - r.p1.time))}`,
  ]
  const w = Math.max(...lines.map((l) => ctx.measureText(l).width)) + 16
  const h = 40
  const cx = (a.x + b.x) / 2
  const top = up ? Math.min(a.y, b.y) - h - 6 : Math.max(a.y, b.y) + 6
  ctx.fillStyle = col
  roundRect(ctx, cx - w / 2, top, w, h, 6)
  ctx.fillStyle = '#ffffff'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(lines[0], cx, top + 13)
  ctx.fillText(lines[1], cx, top + 28)
}

// extend carries the line from a through b on to the edge of the chart.
function extend(a, b, size) {
  const dx = b.x - a.x
  const dy = b.y - a.y
  if (!dx && !dy) return b
  const k = Math.max(size.width, size.height) * 4 / Math.hypot(dx, dy)
  return { x: b.x + dx * k, y: b.y + dy * k }
}

function line(ctx, x1, y1, x2, y2) {
  ctx.beginPath()
  ctx.moveTo(x1, y1)
  ctx.lineTo(x2, y2)
  ctx.stroke()
}

// handle draws a small circle on a drawing's point.
function handle(ctx, p) {
  ctx.save()
  ctx.beginPath()
  ctx.arc(p.x, p.y, 4, 0, Math.PI * 2)
  ctx.fillStyle = '#ffffff'
  ctx.fill()
  ctx.lineWidth = 2
  ctx.stroke()
  ctx.restore()
}

// label draws a price tag ending at x, like the ones on the price axis.
function label(ctx, text, x, y) {
  const w = ctx.measureText(text).width + 12
  ctx.fillStyle = BLUE
  roundRect(ctx, x - w, y - 10, w, 20, 4)
  ctx.fillStyle = '#ffffff'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(text, x - w / 2, y)
}

function roundRect(ctx, x, y, w, h, r) {
  ctx.beginPath()
  ctx.roundRect(x, y, w, h, r)
  ctx.fill()
}

// duration turns seconds into short text: 5400 → "1h 30m", 190800 → "2d 5h".
function duration(s) {
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  if (d) return `${d}d ${h}h`
  if (h) return `${h}h ${m}m`
  return `${m}m`
}
