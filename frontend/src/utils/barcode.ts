// Code39 条码绘制（规则依据 ISO/IEC 16388 / ANSI AIM BC1）
// 每个字符 9 个元素：5 条（2 条宽）+ 4 空（1 个宽），字符间以窄空分隔，首尾加 * 起止符

// 两个宽条的位权为 1,2,4,7,0（0 表示 4+7），其和 1-10 编码数字
const BAR_PATTERNS: Record<number, string> = {
  1: '10000', 2: '01000', 3: '11000', 4: '00100', 5: '10100',
  6: '01100', 7: '00010', 8: '10010', 9: '01010', 10: '00110',
}
// 单个宽空位置决定分组：+30, +0, +10, +20
const GROUP_SPACES: Record<number, string> = {
  30: '1000', 0: '0100', 10: '0010', 20: '0001',
}

function patternFrom(barsValue: number, group: number): string {
  const bars = BAR_PATTERNS[barsValue]
  const spaces = GROUP_SPACES[group]
  let out = ''
  for (let i = 0; i < 4; i++) out += bars[i] + spaces[i]
  return out + bars[4]
}

// $ / / + % 五条全窄，三个宽空，单个窄空分别在第 1-4 位
const SYMBOL_NARROW_SPACE: Record<string, number> = { $: 0, '/': 1, '+': 2, '%': 3 }

const CHAR_PATTERNS: Record<string, string> = {}
for (let i = 1; i <= 9; i++) CHAR_PATTERNS[String(i)] = patternFrom(i, 0)
CHAR_PATTERNS['0'] = patternFrom(10, 0)
'ABCDEFGHIJ'.split('').forEach((ch, i) => { CHAR_PATTERNS[ch] = patternFrom(i + 1, 10) })
'KLMNOPQRST'.split('').forEach((ch, i) => { CHAR_PATTERNS[ch] = patternFrom(i + 1, 20) })
'UVWXYZ'.split('').forEach((ch, i) => { CHAR_PATTERNS[ch] = patternFrom(i + 1, 30) })
CHAR_PATTERNS['-'] = patternFrom(7, 30)
CHAR_PATTERNS['.'] = patternFrom(8, 30)
CHAR_PATTERNS[' '] = patternFrom(9, 30)
CHAR_PATTERNS['*'] = patternFrom(10, 30)
for (const [ch, pos] of Object.entries(SYMBOL_NARROW_SPACE)) {
  let spaces = ''
  for (let i = 0; i < 4; i++) spaces += i === pos ? '0' : '1'
  let p = ''
  for (let i = 0; i < 4; i++) p += '0' + spaces[i]
  CHAR_PATTERNS[ch] = p + '0'
}

const SUPPORTED = new Set(Object.keys(CHAR_PATTERNS))

/** 过滤出可编码字符（大写字母/数字及 - . 空格 $ / + %） */
export function sanitizeCode39(text: string): string {
  return text.toUpperCase().split('').filter((ch) => SUPPORTED.has(ch)).join('')
}

export function isCode39Supported(text: string): boolean {
  return sanitizeCode39(text) === text.toUpperCase() && text.length > 0
}

/**
 * 在 canvas 上绘制 Code39 条码
 * @param text 原始内容（自动转大写并过滤不支持字符）
 * @param options.height 条码高度(px)
 * @param options.narrow 窄元素宽度(px)
 * @returns 是否绘制成功（内容为空返回 false）
 */
export function drawCode39(
  canvas: HTMLCanvasElement,
  text: string,
  options: { height?: number; narrow?: number } = {}
): boolean {
  const content = sanitizeCode39(text)
  if (!content) return false

  const narrow = options.narrow ?? 2
  const height = options.height ?? 60
  const wide = narrow * 3
  const quiet = narrow * 10

  const full = `*${content}*`
  let units = 0
  for (const ch of full) {
    units += CHAR_PATTERNS[ch].split('').reduce((s, b) => s + (b === '1' ? 3 : 1), 0)
    units += 1 // 字符间窄空
  }
  units -= 1 // 末尾字符后无间隔

  const width = quiet * 2 + units * narrow
  const scale = 2 // 提升清晰度
  canvas.width = width * scale
  canvas.height = height * scale
  canvas.style.width = `${width}px`
  canvas.style.height = `${height}px`

  const ctx = canvas.getContext('2d')
  if (!ctx) return false
  ctx.scale(scale, scale)
  ctx.fillStyle = '#fff'
  ctx.fillRect(0, 0, width, height)
  ctx.fillStyle = '#000'

  let x = quiet
  for (let ci = 0; ci < full.length; ci++) {
    const pattern = CHAR_PATTERNS[full[ci]]
    for (let i = 0; i < 9; i++) {
      const w = pattern[i] === '1' ? wide : narrow
      if (i % 2 === 0) ctx.fillRect(x, 0, w, height) // 条
      x += w
    }
    if (ci < full.length - 1) x += narrow
  }
  return true
}
