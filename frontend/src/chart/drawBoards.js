// drawBoards gives every chart its own drawings: one set per token, Price or
// FDV, and candle size. A line drawn on 1h is not on 4h, and is back when 1h
// is picked again, until it is deleted. In memory only: everything goes when
// the page is reloaded (nothing is saved).
const boards = new Map() // chart key (token, unit, size) → the drawing library's JSON

// swap puts the drawings on the manager dm aside as chart `from`, clears it
// and brings back the drawings of chart `to`, if it has any. With no manager
// (the chart is not built yet) there is nothing to swap.
export function swap(dm, from, to) {
  if (!dm) return
  if (from) dm.drawings().length ? boards.set(from, dm.exportJSON()) : boards.delete(from)
  dm.clear()
  if (to && boards.has(to)) dm.importJSON(boards.get(to))
}

// kept is the drawings put aside for chart key, or null.
export const kept = (key) => boards.get(key) ?? null

// forget drops every chart's drawings (used by the tests).
export const forget = () => boards.clear()
