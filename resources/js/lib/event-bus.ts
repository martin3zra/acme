
type EventKey = string | symbol
type EventHandler<T = unknown> = (payload: T) => void
type EventMap = Record<EventKey, EventHandler>


interface EventBus<T extends EventMap> {
  // for a subscriber to listen (subscribe) to an event and register its event handler.
  on<Key extends keyof T>(key: Key, handler: T[Key]): () => void;
  // for a subscriber to remove (unsubscribe) an event and its event handler.
  off<Key extends keyof T>(key: Key, handler: T[Key]): void;
  // for a publisher to send an event to the event bus.
  emit<Key extends keyof T>(key: Key, ...payload: Parameters<T[Key]>): void;
  // for a subscriber to listen to an event only once.
  once<Key extends keyof T>(key: Key, handler: T[Key]): void;
}

type Bus<E> = Record<keyof E, E[keyof E][]>

export function eventbus<E extends EventMap>(config?: {
  onError: (...params: unknown[]) => void
}): EventBus<E> {
  const bus: Partial<Bus<E>> = {}

  const on: EventBus<E>['on'] = (key, handler) => {
    if (bus[key] === undefined)  {
      bus[key] = []
    }
    bus[key]?.push(handler)

    return () => {
      off(key, handler)
    }
  }

  const off: EventBus<E>['off'] = (key, hadler) => {
    const index = bus[key]?.indexOf(hadler) ?? -1
    bus[key]?.splice(index >>> 0,1)
  }

  const emit: EventBus<E>['emit'] = (key, payload) => {
    bus[key]?.forEach((fn) => {
      try {
        fn(payload)
      } catch (e) {
        config?.onError(e)
      }
    })
  }

  const once: EventBus<E>['once'] = (key, handler) => {
    const handleOnce = (payload: Parameters<typeof handler>) => {
      handler(payload)
      off(key, handleOnce as typeof handler)
    }

    on(key, handleOnce as typeof handler)
  }

  return { on, once, off, emit }
}

// Sample
// interface MyBus {
//   'on-event-1': (payload: { data: unknow }) => void
// }

// const myBus = eventbus<MyBus>()