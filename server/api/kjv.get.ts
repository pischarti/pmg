import { defineEventHandler, getQuery } from 'h3'
import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'

type Bible = Record<string, Record<string, Record<string, string>>>

let kjvCache: Bible | null = null

async function loadKjv(): Promise<Bible> {
  if (kjvCache) return kjvCache
  const filePath = resolve(process.cwd(), 'content', 'KJV_bible.json')
  const raw = await readFile(filePath, 'utf-8')
  kjvCache = JSON.parse(raw) as Bible
  return kjvCache
}

const bookMap: Record<string, string> = {
  // OT
  'Gen': 'Genesis', 'Exod': 'Exodus', 'Lev': 'Leviticus', 'Num': 'Numbers', 'Deut': 'Deuteronomy',
  'Josh': 'Joshua', 'Judg': 'Judges', 'Ruth': 'Ruth', '1Sam': '1 Samuel', '2Sam': '2 Samuel',
  '1Kgs': '1 Kings', '2Kgs': '2 Kings', '1Chr': '1 Chronicles', '2Chr': '2 Chronicles',
  'Ezra': 'Ezra', 'Neh': 'Nehemiah', 'Esth': 'Esther', 'Job': 'Job', 'Ps': 'Psalm', 'Prov': 'Proverbs',
  'Eccl': 'Ecclesiastes', 'Song': 'Song of Solomon', 'Isa': 'Isaiah', 'Jer': 'Jeremiah', 'Lam': 'Lamentations',
  'Ezek': 'Ezekiel', 'Dan': 'Daniel', 'Hos': 'Hosea', 'Joel': 'Joel', 'Amos': 'Amos', 'Obad': 'Obadiah',
  'Jonah': 'Jonah', 'Mic': 'Micah', 'Nah': 'Nahum', 'Hab': 'Habakkuk', 'Zeph': 'Zephaniah', 'Hag': 'Haggai',
  'Zech': 'Zechariah', 'Mal': 'Malachi',
  // NT
  'Matt': 'Matthew', 'Mark': 'Mark', 'Luke': 'Luke', 'John': 'John', 'Acts': 'Acts', 'Rom': 'Romans',
  '1Cor': '1 Corinthians', '2Cor': '2 Corinthians', 'Gal': 'Galatians', 'Eph': 'Ephesians',
  'Phil': 'Philippians', 'Col': 'Colossians', '1Thess': '1 Thessalonians', '2Thess': '2 Thessalonians',
  '1Tim': '1 Timothy', '2Tim': '2 Timothy', 'Titus': 'Titus', 'Phlm': 'Philemon', 'Heb': 'Hebrews',
  'Jas': 'James', '1Pet': '1 Peter', '2Pet': '2 Peter', '1John': '1 John', '2John': '2 John',
  '3John': '3 John', 'Jude': 'Jude', 'Rev': 'Revelation'
}

interface RefPoint {
  book: string
  chapter: number
  verse: number
}

function parsePoint(point: string): RefPoint | null {
  // Examples: Ps.19.1  Rom.1.20  1Cor.2.13
  const parts = point.split('.')
  if (parts.length !== 3) return null
  const [abbr, ch, vs] = parts as [string, string, string]
  const book = bookMap[abbr as keyof typeof bookMap]
  if (!book) return null
  const chapter = Number(ch)
  const verse = Number(vs)
  if (!Number.isFinite(chapter) || !Number.isFinite(verse)) return null
  return { book, chapter, verse }
}

function collectVerses(bible: Bible, start: RefPoint, end?: RefPoint): { ref: string, text: string }[] {
  const results: { ref: string, text: string }[] = []
  const endPoint = end ?? start

  if (start.book !== endPoint.book) {
    // Fallback: return just the two endpoints if books differ
    const first = bible[start.book]?.[String(start.chapter)]?.[String(start.verse)]
    if (first) results.push({ ref: `${start.book} ${start.chapter}:${start.verse}`, text: first })
    const second = bible[endPoint.book]?.[String(endPoint.chapter)]?.[String(endPoint.verse)]
    if (second) results.push({ ref: `${endPoint.book} ${endPoint.chapter}:${endPoint.verse}`, text: second })
    return results
  }

  const book = start.book
  if (start.chapter === endPoint.chapter) {
    for (let v = start.verse; v <= endPoint.verse; v++) {
      const text = bible[book]?.[String(start.chapter)]?.[String(v)]
      if (text) results.push({ ref: `${book} ${start.chapter}:${v}`, text })
    }
    return results
  }

  // Multi-chapter range
  for (let ch = start.chapter; ch <= endPoint.chapter; ch++) {
    const chapterObj = bible[book]?.[String(ch)]
    if (!chapterObj) continue
    const verses = Object.keys(chapterObj).map(n => Number(n)).sort((a, b) => a - b)
    if (!verses.length) continue

    // Ensure we have valid verse numbers
    const firstVerse = verses[0]
    const lastVerse = verses[verses.length - 1]
    if (firstVerse === undefined || lastVerse === undefined) continue

    const from = ch === start.chapter ? start.verse : firstVerse
    const to = ch === endPoint.chapter ? endPoint.verse : lastVerse

    for (let v = from; v <= to; v++) {
      const text = chapterObj[String(v)]
      if (text) results.push({ ref: `${book} ${ch}:${v}`, text })
    }
  }

  return results
}

export default defineEventHandler(async (event) => {
  const { q } = getQuery(event)
  if (!q || typeof q !== 'string') {
    return { error: 'Missing query parameter q' }
  }

  const bible = await loadKjv()
  const parts = q.split('-')
  const rawStart = parts[0]
  const rawEnd = parts[1]

  if (!rawStart) {
    return { error: 'Invalid reference format' }
  }

  const start = parsePoint(rawStart)
  const end = rawEnd ? parsePoint(rawEnd) : undefined

  if (!start || (rawEnd && !end)) {
    return { error: 'Invalid reference format' }
  }

  const verses = collectVerses(bible, start, end!)
  return { verses }
})
