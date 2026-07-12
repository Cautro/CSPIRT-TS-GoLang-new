import {useNavigate} from "react-router-dom";
import type {ParallelType} from "../../../../shared/entities/class/types/class_types.ts";
import {useParallels} from "../../hooks/use_parallels.ts";
import {ParallelCard} from "../../../../shared/ui/cards/parallel_card.tsx";

export function ParallelsWidget() {
    const navigate = useNavigate();
    const {data, isLoading, error} = useParallels();
    const classes = data as ParallelType[];

    return (
        <>
            {isLoading && (
                <div className="grid grid--3">
                    <div className="skeleton" style={{ height: 160 }} />
                    <div className="skeleton" style={{ height: 160 }} />
                    <div className="skeleton" style={{ height: 160 }} />
                </div>
            )}

            {error && !isLoading && (
                <div className="alert alert--danger mb-4">{error.message}</div>
            )}

            {classes && !isLoading && !error && classes.length > 0 ? (
                <div className="class-list">
                    {classes.map((item) => (
                        <ParallelCard key={item.Id} item={item} onClick={() => {
                            navigate(`/parallel/${item.Id}?name=${item.Name}`);
                        }}/>
                    ))}
                </div>
            ) : (
                !isLoading && !error && (classes?.length === 0 || !classes) && <div className="empty-state">
                    <h2 className="empty-state__title">Параллели не найдены</h2>
                    <p className="empty-state__text">
                        В системе пока нет доступных параллелей
                    </p>
                </div>
            )}
        </>
    );
} 