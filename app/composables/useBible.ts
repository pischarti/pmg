export const useBible = () => {
  // Function to get bible verses from the API
  const getBibleVerses = async (reference: string): Promise<string> => {
    try {
      // Use the existing KJV API endpoint
      const response = await $fetch<{ verses?: Array<{ ref: string, text: string }>, error?: string }>('/api/kjv', {
        query: { q: reference }
      })

      if (response.verses && Array.isArray(response.verses)) {
        return response.verses
          .map((verse: { ref: string, text: string }) => {
            // Extract verse number from reference (e.g., "Psalms 19:7" -> "7")
            const verseMatch = verse.ref.match(/:(\d+)$/)
            const verseNumber = verseMatch ? verseMatch[1] : '?'
            return `[${verseNumber}] ${verse.text}`
          })
          .join('\n\n')
      }

      return reference
    } catch (error) {
      console.error('Error fetching bible verses:', error)
      return reference
    }
  }

  return {
    getBibleVerses
  }
}
