import { eventbus } from '@/lib/event-bus';

export const messageChannel = eventbus<{
  onSend: (payload: unknown) => void;
}>();

/*
useEffect(() => {
// subscribe to events when mounting
const unsubscribeOnSend = mapEventChannel.on('onSend', () => {
    console.log('on map idle.')
})

// unsubscribe events when unmounting
return () => {
    unsubscribeOnSend()
}
}, [])
*/
