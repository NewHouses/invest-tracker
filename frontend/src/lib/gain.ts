export type GainColor = 'green' | 'red' | 'yellow'

export function gainColor(gain: number): GainColor {
  if (gain > 0) return 'green'
  if (gain < 0) return 'red'
  return 'yellow'
}
