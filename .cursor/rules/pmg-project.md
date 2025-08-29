# PMG Project Rules

## Project Overview
This is a Nuxt.js application with Vue.js components, TypeScript, and content management. The project includes:
- Nuxt.js 3 framework with Vue 3
- TypeScript for type safety
- Content management with markdown and YAML files
- Bible study content (KJV Bible, WCF study materials)
- Blog functionality
- Authentication and user management
- Supabase integration
- Nix development environment

## Code Style & Standards

### TypeScript
- Use strict TypeScript configuration
- Prefer explicit types over `any`
- Use interfaces for object shapes
- Use enums for constants
- Prefer `const` over `let` when possible
- Use optional chaining (`?.`) and nullish coalescing (`??`)

### Vue.js
- Use Composition API with `<script setup>` syntax
- Use TypeScript for component props and emits
- Follow Vue 3 best practices
- Use `defineProps` and `defineEmits` with proper typing
- Prefer reactive refs over reactive objects when possible
- Use `computed` for derived state

### Nuxt.js
- Follow Nuxt 3 conventions
- Use auto-imports when possible
- Place components in `components/` directory
- Use layouts for consistent page structure
- Follow the file-based routing convention

### File Organization
- Components go in `components/` directory
- Pages go in `pages/` directory
- Composables go in `composables/` directory
- Types go in `types/` directory
- Content goes in `content/` directory
- Server API endpoints go in `server/api/` directory

## Naming Conventions

### Files & Directories
- Use kebab-case for file and directory names
- Use PascalCase for Vue component names
- Use camelCase for TypeScript files (utilities, composables)
- Use kebab-case for content files

### Variables & Functions
- Use camelCase for variables and functions
- Use PascalCase for classes and interfaces
- Use UPPER_SNAKE_CASE for constants
- Use descriptive names that explain purpose

## Content Management

### Markdown Files
- Use frontmatter for metadata
- Follow consistent heading structure
- Use relative links when possible
- Include proper alt text for images

### YAML Files
- Use consistent indentation (2 spaces)
- Use descriptive keys
- Group related data logically
- Include comments for complex structures

## API & Data

### Server Endpoints
- Use RESTful naming conventions
- Include proper error handling
- Use TypeScript interfaces for request/response types
- Validate input data

### Database (Supabase)
- Use proper table naming conventions
- Include proper indexes
- Use Row Level Security (RLS) when appropriate
- Handle errors gracefully

## Performance & Best Practices

### Vue Components
- Use `v-memo` for expensive computations
- Lazy load components when appropriate
- Use `shallowRef` for large objects that don't need deep reactivity
- Avoid unnecessary re-renders

### TypeScript
- Use proper type guards
- Avoid type assertions when possible
- Use utility types (Partial, Pick, Omit, etc.)
- Leverage discriminated unions for complex state

## Testing & Quality

### Code Quality
- Use ESLint rules consistently
- Follow Prettier formatting
- Write self-documenting code
- Include JSDoc comments for complex functions

### Error Handling
- Use proper error boundaries
- Log errors appropriately
- Provide user-friendly error messages
- Handle edge cases gracefully

## Development Workflow

### Git
- Use conventional commit messages
- Create feature branches for new work
- Keep commits atomic and focused
- Use descriptive branch names

### Nix Environment
- The project uses Nix for development environment
- All dependencies are managed through flake.nix
- Use `nix develop` to enter the development shell
- Tools like Node.js, pnpm, Go, Python are available

## Special Considerations

### Bible Study Content
- Maintain accuracy in biblical references
- Use consistent formatting for scripture citations
- Handle multiple translations appropriately
- Respect theological content integrity

### Accessibility
- Include proper ARIA labels
- Ensure keyboard navigation works
- Use semantic HTML elements
- Test with screen readers

### Internationalization
- Consider future i18n needs
- Use consistent date/time formatting
- Handle different number formats
- Plan for RTL language support if needed

## Common Patterns

### State Management
- Use composables for shared state
- Prefer local component state when possible
- Use provide/inject for deep component trees
- Consider Pinia for complex state management

### Form Handling
- Use v-model with proper validation
- Implement proper error states
- Use consistent form patterns across components
- Handle form submission gracefully

### Navigation
- Use Nuxt's built-in navigation
- Implement proper loading states
- Handle navigation errors
- Use breadcrumbs for complex navigation

Remember: This is a religious/educational application, so maintain appropriate tone and accuracy in all content and functionality.
