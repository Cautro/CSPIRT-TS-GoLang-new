import {useEffect, useState} from "react";
import {useClasses} from "../../../../features/class/hooks/use_classes.ts";
import {ScheduleWidget} from "../../../../features/schedule/ui/components/schedule_widget.tsx";
import {PageHeader} from "../../../../shared/ui/other/page_header.tsx";
import {useLogout} from "../../../../features/auth/hooks/use_logout.ts";
import {CurrentPublicList} from "../../context/public_context.ts";
import {ParallelsWidget} from "../../../../features/class/ui/components/parallels_widget.tsx";
import {Settings} from "lucide-react";
import {useNavigate} from "react-router-dom";

export function PublicDashboard() {
    const navigate = useNavigate();
    
    const {data: classes} = useClasses()
    const [classId, setClassId] = useState(1)
    const [className, setClassName] = useState("")
    const logout = useLogout()

    function sleep(ms: number) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    useEffect(() => {
        if (!classes || classes.length === 0) return;
        
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setClassId(classes[0].Id ?? 0);
        setClassName(classes[0].Name ?? "")

        let cancelled = false;

        const run = async () => {
            let i = 0;
            while (!cancelled) { 
                setClassId(classes[i].Id ?? 0);
                setClassName(classes[i].Name ?? "");

                await sleep(20000);

                i++;
                if (i >= classes.length) {
                    i = 0;
                }
            }
        }
        
        run()

        return () => {
            cancelled = true;
        };
        
    }, [classes])
    
    return (
        <main className="main"> 
            <section className="page">
                
                <PageHeader 
                    title={`${CurrentPublicList === "schedule" ? `Расписание для ${className} класса` : CurrentPublicList === "parallels" ? "Просматривайте список параллелей школы" : ""}`}
                    actions={
                        <>
                            <button
                                className="app-drawer-button"
                                type="button"
                                onClick={() => navigate("/public/settings")}
                                aria-label="Перейти в профиль"
                            >
                                <Settings size={22} />
                            </button>
                            <button
                                className="btn btn--danger"
                                type="button"
                                onClick={async () => await logout.mutateAsync()}
                                aria-label="Выйти из аккаунта"
                            >
                                Выйти
                            </button>
                        </>
                    }
                />
                
                <div className="page-spacer"></div>

                {CurrentPublicList === "schedule" && (
                    <ScheduleWidget id={classId} name={""} key={classId}/>
                )}

                {CurrentPublicList === "parallels" && (
                    <ParallelsWidget/>
                )}
            </section>
        </main>
    )
}