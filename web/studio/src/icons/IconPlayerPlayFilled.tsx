import { FC } from "react";
import { IconProps } from "./types";

export const IconPlayerPlayFilled: FC<IconProps> = ({
  width = 24,
  height = 24,
  stroke = 1.5,
  fill,
  style,
  className,
}) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={width}
      height={height}
      viewBox="0 0 24 24"
      fill={fill}
      strokeWidth={stroke}
      className={className}
      style={style}
    >
      <path stroke="none" d="M0 0h24v24H0z" fill="none" />
      <path d="M6 4v16a1 1 0 0 0 1.524 .852l13 -8a1 1 0 0 0 0 -1.704l-13 -8a1 1 0 0 0 -1.524 .852z" />
    </svg>
  );
};
