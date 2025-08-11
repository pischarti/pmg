import { defineCollection, z } from '@nuxt/content'

const variantEnum = z.enum([
  'solid',
  'outline',
  'subtle',
  'soft',
  'ghost',
  'link'
])

const colorEnum = z.enum([
  'primary',
  'secondary',
  'neutral',
  'error',
  'warning',
  'success',
  'info'
])

const sizeEnum = z.enum(['xs', 'sm', 'md', 'lg', 'xl'])

const orientationEnum = z.enum(['vertical', 'horizontal'])

const createBaseSchema = () =>
  z.object({
    title: z.string().nonempty(),
    description: z.string().nonempty()
  })

const createFeatureItemSchema = () =>
  createBaseSchema().extend({
    icon: z.string().nonempty().editor({ input: 'icon' })
  })

const createLinkSchema = () =>
  z.object({
    label: z.string().nonempty(),
    to: z.string().nonempty(),
    icon: z.string().optional().editor({ input: 'icon' }),
    size: sizeEnum.optional(),
    trailing: z.boolean().optional(),
    target: z.string().optional(),
    color: colorEnum.optional(),
    variant: variantEnum.optional()
  })

const createImageSchema = () =>
  z.object({
    src: z.string().nonempty().editor({ input: 'media' }),
    alt: z.string().optional(),
    loading: z.string().optional(),
    srcset: z.string().optional()
  })

export const collections = {
  wcf: defineCollection({
    type: 'data',
    source: 'wcf.yml',
    schema: z.object({
      Metadata: z.object({
        AlternativeTitles: z.array(z.string()),
        Authors: z.array(z.string()),
        CreedFormat: z.string(),
        Location: z.string(),
        OriginStory: z.string(),
        OriginalLanguage: z.string(),
        SourceAttribution: z.string(),
        SourceUrl: z.string(),
        Title: z.string(),
        Year: z.string()
      }),
      Data: z.array(
        z.object({
          Chapter: z.string(),
          Sections: z.array(
            z.object({
              Content: z.string(),
              ContentWithProofs: z.string().optional(),
              Proofs: z
                .array(
                  z.object({
                    Id: z.number(),
                    References: z.array(z.string())
                  })
                )
                .optional()
            })
          )
        })
      )
    })
  }),
  authors: defineCollection({
    type: 'data',
    source: 'authors/**.yml',
    schema: z.object({
      name: z.string(),
      avatar: z.string(),
      url: z.string()
    })
  }),
  docs: defineCollection({
    type: 'page',
    source: '1.docs/**/*'
  }),
  posts: defineCollection({
    type: 'page',
    source: '3.blog/**/*',
    schema: z.object({
      image: z.object({
        src: z.string().nonempty().editor({ input: 'media' })
      }),
      authors: z.array(
        z.object({
          name: z.string().nonempty(),
          to: z.string().nonempty(),
          avatar: z.object({
            src: z.string().nonempty().editor({ input: 'media' })
          })
        })
      ),
      date: z.date(),
      badge: z.object({ label: z.string().nonempty() })
    })
  }),
  index: defineCollection({
    source: '0.index.yml',
    type: 'page',
    schema: z.object({
      hero: z.object({
        links: z.array(createLinkSchema())
      }),
      sections: z.array(
        createBaseSchema().extend({
          id: z.string().nonempty(),
          orientation: orientationEnum.optional(),
          reverse: z.boolean().optional(),
          features: z.array(createFeatureItemSchema())
        })
      ),
      features: createBaseSchema().extend({
        items: z.array(createFeatureItemSchema())
      }),
      testimonials: createBaseSchema().extend({
        headline: z.string().optional(),
        items: z.array(
          z.object({
            quote: z.string().nonempty(),
            user: z.object({
              name: z.string().nonempty(),
              description: z.string().nonempty(),
              to: z.string().nonempty(),
              target: z.string().nonempty(),
              avatar: createImageSchema()
            })
          })
        )
      }),
      cta: createBaseSchema().extend({
        links: z.array(createLinkSchema())
      })
    })
  }),
  pricing: defineCollection({
    source: '2.pricing.yml',
    type: 'page',
    schema: z.object({
      plans: z.array(
        z.object({
          title: z.string().nonempty(),
          description: z.string().nonempty(),
          price: z.object({
            month: z.string().nonempty(),
            year: z.string().nonempty()
          }),
          billing_period: z.string().nonempty(),
          billing_cycle: z.string().nonempty(),
          button: createLinkSchema(),
          features: z.array(z.string().nonempty()),
          highlight: z.boolean().optional()
        })
      ),
      logos: z.object({
        title: z.string().nonempty(),
        icons: z.array(z.string())
      }),
      faq: createBaseSchema().extend({
        items: z.array(
          z.object({
            label: z.string().nonempty(),
            content: z.string().nonempty(),
            defaultOpen: z.boolean().optional()
          })
        )
      })
    })
  }),
  kjv: defineCollection({
    type: 'data',
    source: 'kjv/*.yml',
    schema: z.object({})
  }),
  blog: defineCollection({
    source: '3.blog.yml',
    type: 'page'
  })
}
