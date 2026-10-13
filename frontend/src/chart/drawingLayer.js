// DrawingLayer is a Lightweight Charts plugin that draws the user's drawings
// (trend lines, horizontal lines, rectangles, Fibonacci levels, text) and the
// ruler on top of the price chart. Drawings live only in the browser's memory.
//
// A point is { time, price }: time is a bar's unix time, price is in ADA.
import { formatPrice } from '../format.js'

const BLUE = '#2962ff'
const BULL = '#16c784'
const BEAR = '#ea3943'
const FIB_LEVELS = [0, 0.236, 0.382, 0.5, 0.618, 0.786, 1]
const FONT = '12px Inter, system-ui, sans-serif'

export class DrawingLayer {
  constructor() {
    this.drawings = [] // finished drawings
    this.draft = null // the drawing being placed right now
    this.ruler = null // the measurement, if any
    this.hidden = false
    this.t0 = 0 // time of the first bar
    this.step = 1 // seconds between bars
    this._chart = null
    this._series = null
    this._requestUpdate = null
    const layer = this
    this._view = { zOrder: () => 'top', renderer: () => ({ draw: (target) => layer._draw(target) }) }
  }

  // attached is called by the chart when the layer is added to a series.
  attached({ chart, series, requestUpdate }) {
    this._chart = chart
    this._series = series
    this._requestUpdate = requestUpdate
  }

  // detached is called by the chart when the layer is removed from a series.
  detached() {
    this._chart = this._series = this._requestUpdate = null
  }

  paneViews() {
    return [this._view]
  }

  // setBars tells the layer where the bars are, so a time can be turned into
  // an x position even when it is scrolled out of view.
  setBars(t0, step) {
    this.t0 = t0
    this.step = step || 1
  }

  // update asks the chart to redraw the layer.
  update() {
    this._requestUpdate?.()
  }

  // x is the horizontal position of a time, in pixels.
  x(time) {
    return this._chart?.timeScale().logicalToCoordinate((time - this.t0) / this.step) ?? null
  }

  // y is the vertical position of a price, in pixels.
  y(price) {
    return this._series?.priceToCoordinate(price) ?? null
  }

  _draw(target) {
    if (!this._chart || !this._series) return
    target.useMediaCoordinateSpace(({ context: ctx, mediaSize }) => {
      ctx.save()
      ctx.font = FONT
      ctx.lineCap = 'round'
      if (!this.hidden) {
        for (const d of this.drawings) this._shape(ctx, d, mediaSize, false)
        if (this.draft && this.draft.type !== 'ruler') this._shape(ctx, this.draft, mediaSize, true)
      }
      const ruler = this.draft?.type === 'ruler' ? this.draft : this.ruler
      if (ruler) this._ruler(ctx, ruler)
      ctx.restore()
    })
  }

  // _shape draws one drawing. A draft is drawn the same way, with its handles.
  _shape(ctx, d, size, draft) {
    const a = this._xy(d.p1)
    const b = d.p2 ? this._xy(d.p2) : null
    if (!a) return
    ctx.strokeStyle = BLUE
    ctx.fillStyle = BLUE
    ctx.lineWidth = 2
    switch (d.type) {
      case 'trend':
        if (!b) return
        line(ctx, a.x, a.y, b.x, b.y)
        handle(ctx, a)
        handle(ctx, b)
        break
      case 'hline':
        line(ctx, 0, a.y, size.width, a.y)
        label(ctx, formatPrice(d.p1.price), size.width - 6, a.y, BLUE, 'right')
        break
      case 'rect': {
        if (!b) return
        ctx.fillStyle = 'rgba(41, 98, 255, 0.10)'
        ctx.fillRect(Math.min(a.x, b.x), Math.min(a.y, b.y), Math.abs(b.x - a.x), Math.abs(b.y - a.y))
        ctx.strokeRect(Math.min(a.x, b.x), Math.min(a.y, b.y), Math.abs(b.x - a.x), Math.abs(b.y - a.y))
        if (draft) [a, b].forEach((p) => handle(ctx, p))
        break
      }
      case 'fib':
        if (!b) return
        this._fib(ctx, d, a, b)
        break
      case 'text':
        ctx.font = '600 14px Inter, system-ui, sans-serif'
        ctx.textAlign = 'left'
        ctx.textBaseline = 'middle'
        ctx.lineWidth = 4
        ctx.strokeStyle = '#ffffff'
        ctx.strokeText(d.text, a.x, a.y)
        ctx.fillStyle = '#0d1421'
        ctx.fillText(d.text, a.x, a.y)
        break
    }
  }

  // _fib draws Fibonacci retracement levels between the two points.
  _fib(ctx, d, a, b) {
    const left = Math.min(a.x, b.x)
    const right = Math.max(a.x, b.x)
    ctx.lineWidth = 1
    for (const level of FIB_LEVELS) {
      const price = d.p2.price + (d.p1.price - d.p2.price) * level
      const y = this.y(price)
      if (y == null) continue
      line(ctx, left, y, right, y)
      ctx.fillStyle = BLUE
      ctx.textAlign = 'left'
      ctx.textBaseline = 'bottom'
      ctx.fillText(`${level} (${formatPrice(price)})`, left + 4, y - 3)
    }
    ctx.setLineDash([4, 4])
    line(ctx, a.x, a.y, b.x, b.y)
    ctx.setLineDash([])
  }

  // _ruler shows the price change and time between two points.
  _ruler(ctx, r) {
    const a = this._xy(r.p1)
    const b = this._xy(r.p2)
    if (!a || !b) return
    const up = r.p2.price >= r.p1.price
    const col = up ? BULL : BEAR
    ctx.fillStyle = up ? 'rgba(22, 199, 132, 0.14)' : 'rgba(234, 57, 67, 0.14)'
    ctx.fillRect(Math.min(a.x, b.x), Math.min(a.y, b.y), Math.abs(b.x - a.x), Math.abs(b.y - a.y))
    ctx.strokeStyle = col
    ctx.lineWidth = 1.5
    line(ctx, a.x, a.y, b.x, b.y)

    const pct = r.p1.price > 0 ? (r.p2.price / r.p1.price - 1) * 100 : 0
    const diff = r.p2.price - r.p1.price
    const bars = Math.round((r.p2.time - r.p1.time) / this.step)
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

  _xy(p) {
    if (!p) return null
    const x = this.x(p.time)
    const y = this.y(p.price)
    return x == null || y == null ? null : { x, y }
  }
}

// line draws a straight line from (x1, y1) to (x2, y2).
function line(ctx, x1, y1, x2, y2) {
  ctx.beginPath()
  ctx.moveTo(x1, y1)
  ctx.lineTo(x2, y2)
  ctx.stroke()
}

// handle draws a small circle on a drawing's end point.
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

// label draws a small price tag, like the one on the price axis.
function label(ctx, text, x, y, col, align) {
  const w = ctx.measureText(text).width + 12
  const left = align === 'right' ? x - w : x
  ctx.fillStyle = col
  roundRect(ctx, left, y - 10, w, 20, 4)
  ctx.fillStyle = '#ffffff'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(text, left + w / 2, y)
}

// roundRect fills a rectangle with rounded corners.
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
