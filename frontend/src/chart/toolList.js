// The drawing tools on the chart's toolbar, grouped the way Birdeye's
// (TradingView's) toolbar groups them: each group is one button, and its
// arrow opens a menu with section headings. Only the tools that matter for a
// token chart are listed; the library (lightweight-charts-drawing) has more,
// and adding one is adding its kind to a section here, with a label and an
// icon (toolIcons.js).
//
// A tool's kind is its name in the library (its DrawingKind), except
// 'cursor', which draws nothing.

// GROUPS are the group buttons, top to bottom.
export const GROUPS = [
  { key: 'cursor', label: 'Cursor', sections: [{ title: '', kinds: ['cursor'] }] },
  {
    key: 'lines', label: 'Trend line tools',
    sections: [
      { title: 'Lines', kinds: ['trend-line', 'ray', 'horizontal-line', 'horizontal-ray', 'vertical-line'] },
      { title: 'Channels', kinds: ['parallel-channel'] },
    ],
  },
  {
    key: 'fib', label: 'Fibonacci tools',
    sections: [{ title: 'Fibonacci', kinds: ['fib-retracement', 'trend-based-fib-extension'] }],
  },
  {
    key: 'projection', label: 'Projection and measurement tools',
    sections: [
      { title: 'Projection', kinds: ['long-position', 'short-position'] },
      { title: 'Measurer', kinds: ['date-and-price-range'] },
    ],
  },
  {
    key: 'shapes', label: 'Brushes, arrows and shapes',
    sections: [
      { title: 'Brushes', kinds: ['brush'] },
      { title: 'Arrows', kinds: ['arrow'] },
      { title: 'Shapes', kinds: ['rectangle'] },
    ],
  },
  {
    key: 'text', label: 'Text and notes',
    sections: [{ title: 'Text and notes', kinds: ['text', 'price-label'] }],
  },
]

// LABELS are the tools' names, as the menus and tooltips show them.
export const LABELS = {
  cursor: 'Cross',
  'trend-line': 'Trend line',
  ray: 'Ray',
  'horizontal-line': 'Horizontal line',
  'horizontal-ray': 'Horizontal ray',
  'vertical-line': 'Vertical line',
  'parallel-channel': 'Parallel channel',
  'fib-retracement': 'Fib retracement',
  'trend-based-fib-extension': 'Trend-based Fib extension',
  'long-position': 'Long position',
  'short-position': 'Short position',
  'date-and-price-range': 'Date and price range',
  brush: 'Brush',
  arrow: 'Arrow',
  rectangle: 'Rectangle',
  text: 'Text',
  'price-label': 'Price label',
}

// kindsOf lists every tool of a group, in menu order.
export const kindsOf = (group) => group.sections.flatMap((s) => s.kinds)

// groupOf finds the group a tool belongs to (undefined for none).
export const groupOf = (kind) => GROUPS.find((g) => kindsOf(g).includes(kind))

// DRAWING_KINDS are all the kinds placed with the library (every tool but the cursor).
export const DRAWING_KINDS = GROUPS.flatMap(kindsOf).filter((k) => k !== 'cursor')
