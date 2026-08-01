export const useBranding = () => {
  const config = useRuntimeConfig()
  return {
    name: String(config.public.appName),
    shortName: String(config.public.appShortName),
    logo: '/branding/logo.svg',
    favicon: '/branding/favicon.svg'
  }
}

