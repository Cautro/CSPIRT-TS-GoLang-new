import {useState} from "react";
import {useAuthStore} from "../../../auth/store/auth_store.ts";
import type {UserType} from "../../../../shared/entities/user/types/user_types.ts";
import {RatingChangeUseCase} from "../../models/rating_change_usecase.ts";
import {UseChangeRating} from "../../hooks/use_change_rating.ts";

export const RatingSection = ({user, setFormError = () => {}}: {user: UserType, setFormError?: (value: string | null) => void}) => {
  const role = useAuthStore((state) => state.user?.User.Role);
  const normalizedRole = role?.toLowerCase();
  
  const [ratingReason, setRatingReason] = useState("");
  const [ratingValue, setRatingValue] = useState("");
  const rating = user.Rating ?? 0;
  const ratingPercent = Math.min(Math.max((rating / 5000) * 100, 0), 100);
  const canManageRating = normalizedRole === "owner" || normalizedRole === "admin";
  const ratingLevel = rating < 1500 ? "low" : rating < 3500 ? "medium" : "high";
  const changeRating = UseChangeRating()

  async function handleChangeRating() {
    setFormError(null);

    if (!user) {
      setFormError("Пользователь не загружен");
      return;
    }

    const ratingNumber = Number(ratingValue);

    const form = {
      rating: ratingNumber,
      target: user.Login,
      reason: ratingReason.trim(),
    };

    try {
      const dto= RatingChangeUseCase(form);
      await changeRating.mutateAsync(dto);
      setRatingValue("");
      setRatingReason("");
    } catch (e) {
      setFormError(e instanceof Error ? e.message : "Неизвестная ошибка");
    }
  }
  
  return (
      <section className="card card--padded user-rating-card">
        <div className="section-head">
          <h2 className="section-title">Рейтинг</h2>
          <p className="section-description">
            Текущий социальный рейтинг пользователя.
          </p>
        </div>

        <div className="user-rating-summary">
          <div>
            <div className={`profile-rating-value profile-rating-value--${ratingLevel}`}>
              {rating}
            </div>
            <div className="text-muted">из 5000</div>
          </div>

          <span className={`badge badge--${ratingLevel}`}>
                                    {Math.round(ratingPercent)}%
                                </span>
        </div>

        <div className="rating">
          <div className="rating__bar">
            <div
                className={`rating__fill rating__fill--${ratingLevel}`}
                style={{ width: `${ratingPercent}%` }}
            />
          </div>
        </div>

        {canManageRating && (
            <div className="user-action-form user-action-form--compact">
              <div className="field">
                <label className="field__label" htmlFor="ratingValue">
                  Новое значение рейтинга
                </label>

                <input
                    className="input"
                    id="ratingValue"
                    placeholder="Например: 4200"
                    type="number"
                    min={-5000}
                    max={5000}
                    value={ratingValue}
                    onChange={(event) => setRatingValue(event.target.value)}
                />
              </div>

              <div className="field">
                <label className="field__label" htmlFor="ratingReason">
                  Причина изменения
                </label>

                <textarea
                    id="ratingReason"
                    className="textarea"
                    placeholder="Укажите причину изменения рейтинга"
                    maxLength={500}
                    value={ratingReason}
                    onChange={(event) => setRatingReason(event.target.value)}
                />
              </div>

              <div className="user-action-form__footer">
                <button
                    className="btn btn--primary"
                    type="button"
                    disabled={!ratingValue || !ratingReason.trim()}
                    onClick={() => void handleChangeRating()}
                >
                  Изменить рейтинг
                </button>
              </div>
            </div>
        )}
      </section>
  )
}