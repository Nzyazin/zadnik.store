const production = process.env.NODE_ENV === 'production'
const hash = `${Date.now()}`.substring(0, 8)

export default {
  production,
  hash,
  outputFolder: production ? 'build' : 'dev',
  url: 'https://xn--k1aahclgep4d.xn--p1ai/',
  domain: 'задниксторе.рф'
}
