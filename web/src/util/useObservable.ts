import { useEffect, useState, useMemo } from 'react'
import { Observable, Observer, Subject } from 'rxjs'

export function useObservable<T>(observable: Observable<T>, deps: readonly unknown[]): T | undefined {
    const [error, setError] = useState<any>()
    const [currentValue, setCurrentValue] = useState<T>()

    useEffect(() => {
        setCurrentValue(undefined)
        const subscription = observable.subscribe({ next: setCurrentValue, error: setError })
        return () => subscription.unsubscribe()
    }, deps) // eslint-disable-line react-hooks/exhaustive-deps

    if (error) {
        throw error
    }

    return currentValue
}

export function useEventObservable<T, R>(
    transform: (events: Observable<T>) => Observable<R>,
    deps: readonly unknown[]
): [Observer<T>['next'], R | undefined] {
    const { events, observable } = useMemo(() => {
        const events = new Subject<T>()
        const observable = events.pipe(transform)
        return { events, observable }
    }, deps) // eslint-disable-line react-hooks/exhaustive-deps
    const value = useObservable(observable, deps)
    return [events.next.bind(events), value]
}