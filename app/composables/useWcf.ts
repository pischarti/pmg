import type { Ref } from 'vue'
import { useAsyncData } from '#imports'

interface WcfProof {
  Id: number;
  References: string[];
}

interface WcfSection {
  Content: string;
  ContentWithProofs?: string;
  Proofs?: WcfProof[];
}

interface WcfChapter {
  Chapter: string;
  Sections: WcfSection[];
}

interface WcfMetadata {
  AlternativeTitles: string[];
  Authors: string[];
  CreedFormat: string;
  Location: string;
  OriginStory: string;
  OriginalLanguage: string;
  SourceAttribution: string;
  SourceUrl: string;
  Title: string;
  Year: string;
}

interface WcfData {
  Metadata: WcfMetadata;
  Data: WcfChapter[];
}

export const useWcf = async () => {
  const { data: wcf } = await useAsyncData("wcf", () => {
    return queryCollection("wcf").first()
  })

  return {
    wcf: wcf as Ref<WcfData | null>
  }
}
