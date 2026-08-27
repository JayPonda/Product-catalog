const isProduction = import.meta.env.VITE_APP_ENV === 'production'

function formatMessage(correlationId, level, file, method, message, meta) {
  const ts = new Date().toISOString()
  const cid = correlationId || '-'
  const prefix = `[${ts}] [${cid}] ${level} ${file}.${method} ${message}`
  if (meta && Object.keys(meta).length > 0) {
    const pairs = Object.entries(meta)
      .map(([k, v]) => `${k}=${v}`)
      .join(' ')
    return [prefix, pairs]
  }
  return [prefix]
}

class Logger {
  Debug(file, method, message, meta, correlationId) {
    if (isProduction) return
    console.debug(...formatMessage(correlationId, 'DEBUG', file, method, message, meta))
  }

  Info(file, method, message, meta, correlationId) {
    if (isProduction) return
    console.info(...formatMessage(correlationId, 'INFO', file, method, message, meta))
  }

  Warn(file, method, message, meta, correlationId) {
    if (isProduction) return
    console.warn(...formatMessage(correlationId, 'WARN', file, method, message, meta))
  }

  Error(file, method, message, meta, trace, correlationId) {
    if (isProduction) return
    const args = formatMessage(correlationId, 'ERROR', file, method, message, meta)
    if (trace) args.push(`trace=${trace}`)
    console.error(...args)
  }
}

const logger = new Logger()
export default logger
