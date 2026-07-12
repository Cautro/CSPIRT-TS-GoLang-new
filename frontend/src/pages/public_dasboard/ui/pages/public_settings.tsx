import {PageHeader} from "../../../../shared/ui/other/page_header.tsx";
import {useEffect, useState} from "react";
import {changePublicList, CurrentPublicList, type PublicLists} from "../../context/public_context.ts";
import {TabsSwitcher, type TabsSwitcherItem} from "../../../../shared/ui/other/tabs_switcher.tsx";

export function PublicSettings() {
    const [selected, setSelected] = useState<PublicLists>(CurrentPublicList);
    
    useEffect(() => {
        changePublicList(selected);
    }, [selected]);
    
    const tabs: TabsSwitcherItem<PublicLists>[] = [
        {
            label: "Расписание",
            value: "schedule",
        },
        {
            label: "Параллели",
            value: "parallels",
        }
    ]
    
    return (
        <main className="main">
            <section className="page">
                <PageHeader
                    title={"Настройки публичного вида"}
                    hasBackButton={true}
                />
                
                <div className="page-spacer"></div>
                
                <TabsSwitcher 
                    value={selected}
                    onChange={setSelected}
                    items={tabs}
                />
            </section>
        </main>
    )
}