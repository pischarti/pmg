import type { Ref } from 'vue'
import { useAsyncData } from '#imports'

interface WcfSection {
  Content: string
}

interface WcfChapter {
  Chapter: string
  Sections: WcfSection[]
  Title: string
}

interface WcfMetadata {
  Authors: string[]
  OriginalLanguage: string
  Title: string
  Version: string
  Year: string
}

interface WcfData {
  Metadata: WcfMetadata
  Data: WcfChapter[]
}

export const useWcfDiscussion = async () => {
  const { data: wcfstudy } = await useAsyncData('wcfstudy', () => {
    return queryCollection('wcfstudy').first()
  })

  return {
    wcf: wcfstudy as Ref<WcfData | null>
  }
}
