import dayjs from 'dayjs'
import localizedFormat from 'dayjs/plugin/localizedFormat'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(localizedFormat)
dayjs.extend(relativeTime)

export function formatDate(date: Date | string | number, format = 'lll') {
  return dayjs(date).format(format)
}

export function dateFromNow(date: Date | string | number) {
  return dayjs(date).fromNow()
}
