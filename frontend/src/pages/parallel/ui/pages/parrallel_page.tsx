import {PageHeader} from "../../../../shared/ui/other/page_header.tsx";
import {ParallelClassesWidget} from "../../../../features/class/ui/components/parallel_classes_widget.tsx";
import {useParams, useSearchParams} from "react-router-dom";
import type {BurgerDrawerMenuItem} from "../../../../shared/ui/other/burger_menu.tsx";
import {useAuthStore} from "../../../../features/auth/store/auth_store.ts";
import {useState} from "react";
import {CompleteQuarterModal} from "../../../../features/class/ui/components/complete_quarter_modal.tsx";

export function ParallelPage() {
    const [searchParams] = useSearchParams();

    const role = useAuthStore((state) => state.user?.User.Role);
    const normalizedRole = role?.toLowerCase();
    
    const { id } = useParams<{id: string,}>();
    const parallelId = id ? parseInt(id, 10) : null;
    const name = searchParams.get("name");
    
    const menuItems: BurgerDrawerMenuItem[] = [
        {
            label: "Завершить четверть",
            onClick: () => setCompleteQuarterModalOpen(true),
            hidden: !(normalizedRole === "owner")
        }
    ]
    
    const [isCompleteQuarterModalOpen, setCompleteQuarterModalOpen] = useState(false);
    
    return (
        <main className="main">
            <section className="page">
                
                <PageHeader 
                    title={name}
                    description={`Просматривайте информацию о конкретной параллели`}
                    hasBackButton={true}
                    menuItems={menuItems}
                />
                
                <div className="page-spacer"></div>
                
                <ParallelClassesWidget id={parallelId ?? 0} />
                
            </section>
            
            <CompleteQuarterModal
                isOpen={isCompleteQuarterModalOpen}
                onClose={() => setCompleteQuarterModalOpen(false)}
                parallelId={parallelId ?? 0}
            />
        </main>
    )
}