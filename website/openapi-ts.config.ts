import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  client: '@hey-api/client-fetch',
  input: '../docs/swagger.yaml',
  output: 'src/lib/api',
  plugins: [
    'zod', 
    {
      name: '@hey-api/sdk',
      validator: false,
    },
  ],
});
