export const validPassword = (value: string) => {
  const bytes = new TextEncoder().encode(value).length
  return bytes >= 12 && bytes <= 72 && value.trim() === value
}
