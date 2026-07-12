export type PublicLists = "schedule" | "parallels";

export let CurrentPublicList: PublicLists = "schedule";

export function changePublicList(newList: PublicLists) {CurrentPublicList = newList} 